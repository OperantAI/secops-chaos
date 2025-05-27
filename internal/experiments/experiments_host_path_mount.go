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

type HostPathMountExperimentConfig struct {
	Metadata   ExperimentMetadata `yaml:"metadata"`
	Parameters HostPathMount      `yaml:"parameters"`
}

type HostPathMount struct {
	HostPath HostPath `yaml:"hostPath"`
}

type HostPath struct {
	Path string `yaml:"path"`
}

func (p *HostPathMountExperimentConfig) Type() string {
	return "host-path-mount"
}

func (p *HostPathMountExperimentConfig) Description() string {
	return "Mount a sensitive host filesystem path into a container"
}

func (p *HostPathMountExperimentConfig) Technique() string {
	return categories.MITRE.PrivilegeEscalation.HostPathMount.Technique
}

func (p *HostPathMountExperimentConfig) Tactic() string {
	return categories.MITRE.PrivilegeEscalation.HostPathMount.Tactic
}

func (p *HostPathMountExperimentConfig) Framework() string {
	return string(categories.Mitre)
}

func (p *HostPathMountExperimentConfig) Run(ctx context.Context, experimentConfig *ExperimentConfig) error {
	client, err := k8s.NewClient()
	if err != nil {
		return err
	}
	var hostPathMountExperimentConfig HostPathMountExperimentConfig
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err = yaml.Unmarshal(yamlObj, &hostPathMountExperimentConfig)
	if err != nil {
		return err
	}
	params := hostPathMountExperimentConfig.Parameters
	clientset := client.Clientset
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: hostPathMountExperimentConfig.Metadata.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": hostPathMountExperimentConfig.Metadata.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"experiment": hostPathMountExperimentConfig.Metadata.Name,
						"app":        hostPathMountExperimentConfig.Metadata.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            hostPathMountExperimentConfig.Metadata.Name,
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
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "hostpath-volume",
									MountPath: "/tmp",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "hostpath-volume",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: params.HostPath.Path,
								},
							},
						},
					},
				},
			},
		},
	}
	_, err = clientset.AppsV1().Deployments(hostPathMountExperimentConfig.Metadata.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	return err
}

func (p *HostPathMountExperimentConfig) Verify(ctx context.Context, experimentConfig *ExperimentConfig) (*verifier.Outcome, error) {
	client, err := k8s.NewClient()
	if err != nil {
		return nil, err
	}
	var hostPathMountExperimentConfig HostPathMountExperimentConfig
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err = yaml.Unmarshal(yamlObj, &hostPathMountExperimentConfig)
	if err != nil {
		return nil, err
	}
	params := hostPathMountExperimentConfig.Parameters
	v := verifier.New(
		hostPathMountExperimentConfig.Metadata.Name,
		hostPathMountExperimentConfig.Description(),
		hostPathMountExperimentConfig.Framework(),
		hostPathMountExperimentConfig.Tactic(),
		hostPathMountExperimentConfig.Technique(),
	)
	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", hostPathMountExperimentConfig.Metadata.Name),
	}
	clientset := client.Clientset
	pods, err := clientset.CoreV1().Pods(hostPathMountExperimentConfig.Metadata.Namespace).List(ctx, listOptions)
	if err != nil {
		return nil, err
	}
	if len(pods.Items) == 1 {
		if checkVolumes(pods.Items[0], params.HostPath.Path) {
			v.Success("")
		} else {
			v.Fail("")
		}
	}
	return v.GetOutcome(), nil
}

func checkVolumes(pod corev1.Pod, volumePath string) bool {
	if pod.Status.Phase == "Running" {
		volumes := pod.Spec.Volumes
		for _, v := range volumes {
			if v.HostPath != nil && v.HostPath.Path == volumePath {
				return true
			}
		}
	}
	return false
}

func (p *HostPathMountExperimentConfig) Cleanup(ctx context.Context, experimentConfig *ExperimentConfig) error {
	client, err := k8s.NewClient()
	if err != nil {
		return err
	}
	var hostPathMountExperimentConfig HostPathMountExperimentConfig
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err = yaml.Unmarshal(yamlObj, &hostPathMountExperimentConfig)
	if err != nil {
		return err
	}
	clientset := client.Clientset
	return clientset.AppsV1().Deployments(hostPathMountExperimentConfig.Metadata.Namespace).Delete(ctx, hostPathMountExperimentConfig.Metadata.Name, metav1.DeleteOptions{})
}
