/*
Copyright 2023 Operant AI
*/
package experiments

import (
	"context"
	"fmt"

	"github.com/operantai/secops-chaos/internal/categories"
	"github.com/operantai/secops-chaos/internal/k8s"
	"github.com/operantai/secops-chaos/internal/verifier"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

type ContainerSecretsExperiment struct {
	Metadata   ExperimentMetadata
	Parameters ContainerSecrets
}

type ContainerSecrets struct {
	ConfigMapCheck bool                  `yaml:"configMapCheck"`
	PodEnvCheck    bool                  `yaml:"podEnvCheck"`
	Env            []ContainerSecretsEnv `yaml:"env"`
}

type ContainerSecretsEnv struct {
	EnvKey   string `yaml:"envKey"`
	EnvValue string `yaml:"envValue"`
}

func (p *ContainerSecretsExperiment) Name() string {
	return p.Metadata.Name
}

func (p *ContainerSecretsExperiment) Type() string {
	return "credential-access-container-secrets"
}

func (p *ContainerSecretsExperiment) Description() string {
	return "Add secrets to a config map and within a container's environment variables"
}

func (p *ContainerSecretsExperiment) Technique() string {
	return categories.MITRE.Credentials.ApplicationCredentialsInConfigurationFiles.Technique
}

func (p *ContainerSecretsExperiment) Tactic() string {
	return categories.MITRE.Credentials.ApplicationCredentialsInConfigurationFiles.Tactic
}

func (p *ContainerSecretsExperiment) Framework() string {
	return string(categories.Mitre)
}

func (p *ContainerSecretsExperiment) DependsOn() []string {
	return p.Metadata.DependsOn
}

func (p *ContainerSecretsExperiment) Run(ctx context.Context, client *k8s.Client) error {
	params := p.Parameters
	clientset := client.Clientset
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: p.Metadata.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": p.Metadata.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"experiment": p.Metadata.Name,
						"app":        p.Metadata.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            p.Metadata.Name,
							Image:           "alpine:latest",
							ImagePullPolicy: corev1.PullAlways,
							Command: []string{
								"sh",
								"-c",
								"while true; do :; done",
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 4000,
								},
							},
						},
					},
				},
			},
		},
	}
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: p.Metadata.Name,
			Labels: map[string]string{
				"app":        p.Metadata.Name,
				"experiment": p.Metadata.Name,
			},
		},
	}
	configMap.Data = map[string]string{}
	for _, item := range params.Env {
		deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
			Name:  item.EnvKey,
			Value: item.EnvValue,
		})
		configMap.Data[item.EnvKey] = item.EnvValue
	}
	if params.PodEnvCheck {
		_, err := clientset.AppsV1().Deployments(p.Metadata.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}
	if params.ConfigMapCheck {
		_, err := clientset.CoreV1().ConfigMaps(p.Metadata.Namespace).Create(ctx, configMap, metav1.CreateOptions{})
		return err
	}
	return nil
}

func (p *ContainerSecretsExperiment) Verify(ctx context.Context, client *k8s.Client) (*verifier.Outcome, error) {
	params := p.Parameters
	clientset := client.Clientset
	v := verifier.New(
		p.Metadata.Name,
		p.Description(),
		p.Framework(),
		p.Tactic(),
		p.Technique(),
	)
	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", p.Metadata.Name),
	}
	if params.PodEnvCheck {
		pods, err := clientset.CoreV1().Pods(p.Metadata.Namespace).List(ctx, listOptions)
		if err != nil {
			return nil, err
		}

		if len(pods.Items) == 1 {
			if checkPodForSecrets(pods.Items[0], params.Env) {
				v.Success("PodEnvironmentHasSecrets")
			} else {
				v.Fail("PodEnvironmentHasSecrets")
			}
		}
	}
	if params.ConfigMapCheck {
		configMaps, err := clientset.CoreV1().ConfigMaps(p.Metadata.Namespace).List(ctx, listOptions)
		if err != nil {
			return nil, err
		}
		if len(configMaps.Items) == 1 {
			if checkConfigMapForSecrets(configMaps.Items[0], params.Env) {
				v.Success("ConfigMapHasSecrets")
			} else {
				v.Fail("ConfigMapHasSecrets")
			}
		}
	}
	return v.GetOutcome(), nil
}

func checkPodForSecrets(pod corev1.Pod, envTest []ContainerSecretsEnv) bool {
	if pod.Status.Phase == "Running" && pod.Spec.Containers != nil && len(pod.Spec.Containers) > 0 {
		podEnv := pod.Spec.Containers[0].Env
		for _, item := range podEnv {
			for _, envItem := range envTest {
				if item.Name == envItem.EnvKey {
					return true
				}
				if item.Value == envItem.EnvValue {
					return true
				}
			}
		}
	}
	return false
}

func checkConfigMapForSecrets(configMap corev1.ConfigMap, envTest []ContainerSecretsEnv) bool {
	if configMap.Data != nil {
		cmMap := configMap.Data
		for k, v := range cmMap {
			for _, envItem := range envTest {
				if k == envItem.EnvKey {
					return true
				}
				if v == envItem.EnvValue {
					return true
				}
			}
		}
	}
	return false
}

func (p *ContainerSecretsExperiment) Cleanup(ctx context.Context, client *k8s.Client) error {
	params := p.Parameters
	clientset := client.Clientset
	if params.PodEnvCheck {
		err := clientset.AppsV1().Deployments(p.Metadata.Namespace).Delete(ctx, p.Metadata.Name, metav1.DeleteOptions{})

		if err != nil {
			return err
		}
	}
	if params.ConfigMapCheck {
		return clientset.CoreV1().ConfigMaps(p.Metadata.Namespace).Delete(ctx, p.Metadata.Name, metav1.DeleteOptions{})
	}
	return nil
}
