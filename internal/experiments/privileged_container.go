/*
Copyright 2023 Operant AI
*/
package experiments

import (
	"context"

	"github.com/operantai/experiments-runtime-tool/internal/output"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/pointer"
)

type PrivilegedContainer struct {
	HostPid     bool `yaml:"host_pid"`
	HostNetwork bool `yaml:"host_network"`
	RunAsRoot   bool `yaml:"run_as_root"`
}

func (p *PrivilegedContainer) Name() string {
	return "PrivilegedContainer"
}

func (p *PrivilegedContainer) Run(ctx context.Context, client *kubernetes.Clientset) error {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &config.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": config.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"experiment": config.Name,
						"app":        config.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            config.Name,
							Image:           config.Image,
							ImagePullPolicy: corev1.PullAlways,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: config.Port,
								},
							},
						},
					},
				},
			},
		},
	}
	if p.HostPid {
		deployment.Spec.Template.Spec.HostPID = true
	}
	if p.HostNetwork {
		deployment.Spec.Template.Spec.HostNetwork = true
	}
	if p.RunAsRoot {
		deployment.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
			RunAsUser: pointer.Int64(0),
		}
	}
	output.WriteInfo("Creating experiment: %s", config.Name)
	_, err := client.AppsV1().Deployments(config.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	return err
}

func (p *PrivilegedContainer) Cleanup(ctx context.Context, client *kubernetes.Clientset) error {
	output.WriteInfo("Deleting experiment: %s", config.Name)
	return client.AppsV1().Deployments(config.Namespace).Delete(ctx, config.Name, metav1.DeleteOptions{})
}

var _ Experiment = (*PrivilegedContainer)(nil)
