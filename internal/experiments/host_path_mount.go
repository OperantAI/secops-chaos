/*
Copyright 2023 Operant AI
*/
package experiments

import (
	"context"
	"fmt"
	"github.com/operantai/secops-chaos/internal/categories"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/pointer"
)

type HostPathMountExperimentConfig struct {
	Metadata   ExperimentMetadata `yaml:"metadata"`
	Parameters HostPathMount      `yaml:"parameters"`
}

type HostPathMount struct {
	HostPath HostPath `yaml:"host_path"`
}

type HostPath struct {
	Path string `yaml:"path"`
}

func (p *HostPathMountExperimentConfig) Type() string {
	return "host_path_mount"
}

func (p *HostPathMountExperimentConfig) Description() string {
	return "This experiment attempts to mount a sensitive host filesystem path into a container"
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

func (p *HostPathMountExperimentConfig) Run(ctx context.Context, client *kubernetes.Clientset, experimentConfig *ExperimentConfig) error {
	var hostPathMountExperimentConfig HostPathMountExperimentConfig
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err := yaml.Unmarshal(yamlObj, &hostPathMountExperimentConfig)
	if err != nil {
		return err
	}
	params := hostPathMountExperimentConfig.Parameters
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
	_, err = client.AppsV1().Deployments(hostPathMountExperimentConfig.Metadata.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	return err
}

func (p *HostPathMountExperimentConfig) Verify(ctx context.Context, client *kubernetes.Clientset, experimentConfig *ExperimentConfig) (*Outcome, error) {
	var hostPathMountExperimentConfig HostPathMountExperimentConfig
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err := yaml.Unmarshal(yamlObj, &hostPathMountExperimentConfig)
	if err != nil {
		return nil, err
	}
	params := hostPathMountExperimentConfig.Parameters
	outcome := &Outcome{
		Experiment:  hostPathMountExperimentConfig.Metadata.Name,
		Description: hostPathMountExperimentConfig.Description(),
		Framework:   hostPathMountExperimentConfig.Framework(),
		Tactic:      hostPathMountExperimentConfig.Tactic(),
		Technique:   hostPathMountExperimentConfig.Technique(),
		Success:     false,
	}
	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", hostPathMountExperimentConfig.Metadata.Name),
	}
	pods, err := client.CoreV1().Pods(hostPathMountExperimentConfig.Metadata.Namespace).List(ctx, listOptions)
	if err != nil {
		return nil, err
	}
	if len(pods.Items) == 1 && pods.Items[0].Spec.Volumes[0].HostPath.Path == params.HostPath.Path && pods.Items[0].Status.Phase == "Running" {
		outcome.Success = true
	}
	return outcome, nil
}

func (p *HostPathMountExperimentConfig) Cleanup(ctx context.Context, client *kubernetes.Clientset, experimentConfig *ExperimentConfig) error {
	var hostPathMountExperimentConfig HostPathMountExperimentConfig
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err := yaml.Unmarshal(yamlObj, &hostPathMountExperimentConfig)
	if err != nil {
		return err
	}
	return client.AppsV1().Deployments(hostPathMountExperimentConfig.Metadata.Namespace).Delete(ctx, hostPathMountExperimentConfig.Metadata.Name, metav1.DeleteOptions{})
}