/*
Copyright 2023 Operant AI
*/
package experiments

import (
	"context"
	"fmt"

	"github.com/operantai/woodpecker/internal/categories"
	"github.com/operantai/woodpecker/internal/k8s"
	"github.com/operantai/woodpecker/internal/verifier"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

type ContainerSecretsExperimentConfig struct {
	Metadata   ExperimentMetadata `yaml:"metadata"`
	Parameters ContainerSecrets   `yaml:"parameters"`
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

func (p *ContainerSecretsExperimentConfig) Type() string {
	return "credential-access-container-secrets"
}

func (p *ContainerSecretsExperimentConfig) Description() string {
	return "Add secrets to a config map and within a container's environment variables"
}

func (p *ContainerSecretsExperimentConfig) Technique() string {
	return categories.MITRE.Credentials.ApplicationCredentialsInConfigurationFiles.Technique
}

func (p *ContainerSecretsExperimentConfig) Tactic() string {
	return categories.MITRE.Credentials.ApplicationCredentialsInConfigurationFiles.Tactic
}

func (p *ContainerSecretsExperimentConfig) Framework() string {
	return string(categories.Mitre)
}

func (p *ContainerSecretsExperimentConfig) Run(ctx context.Context, experimentConfig *ExperimentConfig) error {
	client, err := k8s.NewClient()
	if err != nil {
		return err
	}
	var containerSecretsExperimentConfig ContainerSecretsExperimentConfig
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err = yaml.Unmarshal(yamlObj, &containerSecretsExperimentConfig)
	if err != nil {
		return err
	}
	params := containerSecretsExperimentConfig.Parameters
	clientset := client.Clientset
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: containerSecretsExperimentConfig.Metadata.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": containerSecretsExperimentConfig.Metadata.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"experiment": containerSecretsExperimentConfig.Metadata.Name,
						"app":        containerSecretsExperimentConfig.Metadata.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            containerSecretsExperimentConfig.Metadata.Name,
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
			Name: containerSecretsExperimentConfig.Metadata.Name,
			Labels: map[string]string{
				"app":        containerSecretsExperimentConfig.Metadata.Name,
				"experiment": containerSecretsExperimentConfig.Metadata.Name,
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
		_, err = clientset.AppsV1().Deployments(containerSecretsExperimentConfig.Metadata.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}
	if params.ConfigMapCheck {
		_, err = clientset.CoreV1().ConfigMaps(containerSecretsExperimentConfig.Metadata.Namespace).Create(ctx, configMap, metav1.CreateOptions{})
		return err
	}
	return nil
}

func (p *ContainerSecretsExperimentConfig) Verify(ctx context.Context, experimentConfig *ExperimentConfig) (*verifier.Outcome, error) {
	client, err := k8s.NewClient()
	if err != nil {
		return nil, err
	}
	var containerSecretsExperimentConfig ContainerSecretsExperimentConfig
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err = yaml.Unmarshal(yamlObj, &containerSecretsExperimentConfig)
	if err != nil {
		return nil, err
	}
	params := containerSecretsExperimentConfig.Parameters
	clientset := client.Clientset
	v := verifier.New(
		containerSecretsExperimentConfig.Metadata.Name,
		containerSecretsExperimentConfig.Description(),
		containerSecretsExperimentConfig.Framework(),
		containerSecretsExperimentConfig.Tactic(),
		containerSecretsExperimentConfig.Technique(),
	)
	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", containerSecretsExperimentConfig.Metadata.Name),
	}
	if params.PodEnvCheck {
		pods, err := clientset.CoreV1().Pods(containerSecretsExperimentConfig.Metadata.Namespace).List(ctx, listOptions)
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
		configMaps, err := clientset.CoreV1().ConfigMaps(containerSecretsExperimentConfig.Metadata.Namespace).List(ctx, listOptions)
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

func (p *ContainerSecretsExperimentConfig) Cleanup(ctx context.Context, experimentConfig *ExperimentConfig) error {
	client, err := k8s.NewClient()
	if err != nil {
		return err
	}
	var containerSecretsExperimentConfig ContainerSecretsExperimentConfig
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err = yaml.Unmarshal(yamlObj, &containerSecretsExperimentConfig)
	if err != nil {
		return err
	}
	params := containerSecretsExperimentConfig.Parameters
	clientset := client.Clientset
	if params.PodEnvCheck {
		err = clientset.AppsV1().Deployments(containerSecretsExperimentConfig.Metadata.Namespace).Delete(ctx, containerSecretsExperimentConfig.Metadata.Name, metav1.DeleteOptions{})

		if err != nil {
			return err
		}
	}
	if params.ConfigMapCheck {
		return clientset.CoreV1().ConfigMaps(containerSecretsExperimentConfig.Metadata.Namespace).Delete(ctx, containerSecretsExperimentConfig.Metadata.Name, metav1.DeleteOptions{})
	}
	return nil
}
