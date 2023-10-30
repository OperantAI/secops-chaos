/*
Copyright 2023 Operant AI
*/
package experiments

import (
	"context"

	"github.com/mitchellh/mapstructure"

	"github.com/operantai/secops-chaos/internal/categories"
	"github.com/operantai/secops-chaos/internal/verifier"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/pointer"
)

type PrivilegedContainerExperimentConfig struct {
	Metadata   ExperimentMetadata  `yaml:"metadata"`
	Parameters PrivilegedContainer `yaml:"parameters"`
}

// PrivilegedContainer is an experiment that creates a deployment with a privileged container
type PrivilegedContainer struct {
	Privileged  bool `yaml:"privileged"`
	HostPid     bool `yaml:"host_pid"`
	HostNetwork bool `yaml:"host_network"`
	RunAsRoot   bool `yaml:"run_as_root"`
}

func (p *PrivilegedContainerExperimentConfig) Type() string {
	return "privileged_container"
}

func (p *PrivilegedContainerExperimentConfig) Description() string {
	return "This experiment attempts to run a privileged container in a namespace"
}

func (p *PrivilegedContainerExperimentConfig) Technique() string {
	return categories.MITRE.PrivilegeEscalation.PrivilegedContainer.Technique
}

func (p *PrivilegedContainerExperimentConfig) Tactic() string {
	return categories.MITRE.PrivilegeEscalation.PrivilegedContainer.Tactic
}

func (p *PrivilegedContainerExperimentConfig) Framework() string {
	return string(categories.Mitre)
}

func (p *PrivilegedContainerExperimentConfig) Run(ctx context.Context, client *kubernetes.Clientset, experimentConfig *ExperimentConfig) error {
	var privilegedContainerExperimentConfig PrivilegedContainerExperimentConfig
	err := mapstructure.Decode(experimentConfig, &privilegedContainerExperimentConfig)
	if err != nil {
		return err
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: privilegedContainerExperimentConfig.Metadata.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": privilegedContainerExperimentConfig.Metadata.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"experiment": privilegedContainerExperimentConfig.Metadata.Name,
						"app":        privilegedContainerExperimentConfig.Metadata.Name,
					},
				},
				Spec: corev1.PodSpec{
					HostNetwork: privilegedContainerExperimentConfig.Parameters.HostNetwork,
					HostPID:     privilegedContainerExperimentConfig.Parameters.HostPid,
					Containers: []corev1.Container{
						{
							Name:            privilegedContainerExperimentConfig.Metadata.Name,
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
	params := privilegedContainerExperimentConfig.Parameters
	if params.RunAsRoot {
		if deployment.Spec.Template.Spec.Containers[0].SecurityContext == nil {
			deployment.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{
				RunAsUser: pointer.Int64(0),
			}
		} else {
			deployment.Spec.Template.Spec.Containers[0].SecurityContext.RunAsUser = pointer.Int64(0)
		}
	}

	if params.Privileged {
		if deployment.Spec.Template.Spec.Containers[0].SecurityContext == nil {
			deployment.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{
				Privileged: pointer.Bool(true),
			}
		} else {
			deployment.Spec.Template.Spec.Containers[0].SecurityContext.Privileged = pointer.Bool(true)
		}
	}

	_, err = client.AppsV1().Deployments(privilegedContainerExperimentConfig.Metadata.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	return err
}

func (p *PrivilegedContainerExperimentConfig) Verify(ctx context.Context, client *kubernetes.Clientset, experimentConfig *ExperimentConfig) (*verifier.Outcome, error) {
	var privilegedContainerExperimentConfig PrivilegedContainerExperimentConfig
	err := mapstructure.Decode(experimentConfig, &privilegedContainerExperimentConfig)
	if err != nil {
		return nil, err
	}
	deployment, err := client.AppsV1().Deployments(privilegedContainerExperimentConfig.Metadata.Namespace).Get(ctx, privilegedContainerExperimentConfig.Metadata.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	params := privilegedContainerExperimentConfig.Parameters

	verifier := verifier.New(
		privilegedContainerExperimentConfig.Metadata.Name,
		privilegedContainerExperimentConfig.Description(),
		privilegedContainerExperimentConfig.Framework(),
		privilegedContainerExperimentConfig.Tactic(),
		privilegedContainerExperimentConfig.Technique(),
	)

	if params.HostPid {
		if deployment.Spec.Template.Spec.HostPID {
			verifier.Success()
		}
	}

	if params.HostNetwork {
		if deployment.Spec.Template.Spec.HostNetwork {
			verifier.Success()
		}
	}

	if params.RunAsRoot {
		if deployment.Spec.Template.Spec.Containers[0].SecurityContext != nil {
			if deployment.Spec.Template.Spec.Containers[0].SecurityContext.RunAsUser != nil {
				if *deployment.Spec.Template.Spec.Containers[0].SecurityContext.RunAsUser == 0 {
					verifier.Success()
				}
			}
		}
	}

	if params.Privileged {
		if deployment.Spec.Template.Spec.Containers[0].SecurityContext != nil {
			if deployment.Spec.Template.Spec.Containers[0].SecurityContext.Privileged != nil {
				if *deployment.Spec.Template.Spec.Containers[0].SecurityContext.Privileged {
					verifier.Success()
				}
			}
		}
	}

	return verifier.GetOutcome(), nil
}

func (p *PrivilegedContainerExperimentConfig) Cleanup(ctx context.Context, client *kubernetes.Clientset, experimentConfig *ExperimentConfig) error {
	var privilegedContainerExperimentConfig PrivilegedContainerExperimentConfig
	err := mapstructure.Decode(experimentConfig, &privilegedContainerExperimentConfig)
	if err != nil {
		return err
	}
	return client.AppsV1().Deployments(privilegedContainerExperimentConfig.Metadata.Namespace).Delete(ctx, privilegedContainerExperimentConfig.Metadata.Name, metav1.DeleteOptions{})
}
