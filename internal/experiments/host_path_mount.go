/*
Copyright 2023 Operant AI
*/
package experiments

import (
	"context"
	"github.com/mitchellh/mapstructure"
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

func (p *HostPathMountExperimentConfig) Category() string {
	return categories.MITRE.PrivilegeEscalation.HostPathMount.Name
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
	err := mapstructure.Decode(experimentConfig, &hostPathMountExperimentConfig)
	if err != nil {
		return nil, err
	}
	/*deployment, err := client.AppsV1().Deployments(hostPathMountExperimentConfig.Metadata.Namespace).Get(ctx, hostPathMountExperimentConfig.Metadata.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	params := hostPathMountExperimentConfig.Parameters*/
	outcome := &Outcome{
		Experiment: hostPathMountExperimentConfig.Metadata.Name,
		Category:   hostPathMountExperimentConfig.Category(),
		Success:    false,
	}
	return outcome, nil
}

func (p *HostPathMountExperimentConfig) Cleanup(ctx context.Context, client *kubernetes.Clientset, experimentConfig *ExperimentConfig) error {
	var hostPathMountExperimentConfig HostPathMountExperimentConfig
	err := mapstructure.Decode(experimentConfig, &hostPathMountExperimentConfig)
	if err != nil {
		return err
	}
	return client.AppsV1().Deployments(hostPathMountExperimentConfig.Metadata.Namespace).Delete(ctx, hostPathMountExperimentConfig.Metadata.Name, metav1.DeleteOptions{})
}
