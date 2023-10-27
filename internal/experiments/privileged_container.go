/*
Copyright 2023 Operant AI
*/
package experiments

import (
	"context"

	"github.com/operantai/secops-chaos/internal/categories"
	"github.com/operantai/secops-chaos/internal/output"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/pointer"
)

type PrivilegedContainer struct {
	Privileged  bool `yaml:"privileged"`
	HostPid     bool `yaml:"host_pid"`
	HostNetwork bool `yaml:"host_network"`
	RunAsRoot   bool `yaml:"run_as_root"`
}

func (p *PrivilegedContainer) Name() string {
	return "privileged_container"
}

func (p *PrivilegedContainer) Category() string {
	return categories.MITRE.PrivilegeEscalation.PrivilegedContainer.Name
}

func (p *PrivilegedContainer) Run(ctx context.Context, client *kubernetes.Clientset, config *ExperimentConfig) error {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(1),
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
							Image:           "alpine:latest",
							ImagePullPolicy: corev1.PullAlways,
							Command: []string{
								"sleep",
								"1000000",
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
	params := config.Parameters.(PrivilegedContainer)
	if params.HostPid {
		deployment.Spec.Template.Spec.HostPID = true
	}
	if params.HostNetwork {
		deployment.Spec.Template.Spec.HostNetwork = true
	}
	if params.RunAsRoot {
		deployment.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
			RunAsUser: pointer.Int64(0),
		}
	}

	output.WriteInfo("Creating experiment: %s", config.Name)
	_, err := client.AppsV1().Deployments(config.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	return err
}

func (p *PrivilegedContainer) Verify(ctx context.Context, client *kubernetes.Clientset, config *ExperimentConfig) (*Outcome, error) {
	deployment, err := client.AppsV1().Deployments(config.Namespace).Get(ctx, config.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	params := config.Parameters.(PrivilegedContainer)
	outcome := &Outcome{
		Experiment: config.Name,
		Category:   p.Category(),
		Success:    false,
	}

	if params.HostPid {
		if deployment.Spec.Template.Spec.HostPID {
			outcome.Success = true
			return outcome, nil
		}
	}

	if params.HostNetwork {
		if deployment.Spec.Template.Spec.HostNetwork {
			outcome.Success = true
			return outcome, nil
		}
	}

	if params.Privileged {
		if deployment.Spec.Template.Spec.Containers[0].SecurityContext != nil {
			if deployment.Spec.Template.Spec.Containers[0].SecurityContext.Privileged != nil {
				if *deployment.Spec.Template.Spec.Containers[0].SecurityContext.Privileged {
					outcome.Success = true
					return outcome, nil
				}
			}
		}
	}

	return outcome, nil
}

func (p *PrivilegedContainer) Cleanup(ctx context.Context, client *kubernetes.Clientset, config *ExperimentConfig) error {
	return client.AppsV1().Deployments(config.Namespace).Delete(ctx, config.Name, metav1.DeleteOptions{})
}

var _ Experiment = (*PrivilegedContainer)(nil)
