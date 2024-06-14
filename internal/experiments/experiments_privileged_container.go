/*
Copyright 2023 Operant AI
*/
package experiments

import (
	"context"

	"gopkg.in/yaml.v3"

	"github.com/operantai/secops-chaos/internal/categories"
	"github.com/operantai/secops-chaos/internal/k8s"
	"github.com/operantai/secops-chaos/internal/verifier"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/utils/pointer"
)

type PrivilegedContainerExperiment struct {
	*ExperimentConfig
}

// PrivilegedContainer is an experiment that creates a deployment with a privileged container
type PrivilegedContainer struct {
	Image       string   `yaml:"image"`
	Command     []string `yaml:"command"`
	Privileged  bool     `yaml:"privileged"`
	HostPid     bool     `yaml:"hostPid"`
	HostNetwork bool     `yaml:"hostNetwork"`
	RunAsRoot   bool     `yaml:"runAsRoot"`
}

func (p *PrivilegedContainerExperiment) Name() string {
	return p.Metadata.Name
}

func (p *PrivilegedContainerExperiment) Type() string {
	return "privileged-container"
}

func (p *PrivilegedContainerExperiment) Description() string {
	return "Run a privileged container in a namespace"
}

func (p *PrivilegedContainerExperiment) Technique() string {
	return categories.MITRE.PrivilegeEscalation.PrivilegedContainer.Technique
}

func (p *PrivilegedContainerExperiment) Tactic() string {
	return categories.MITRE.PrivilegeEscalation.PrivilegedContainer.Tactic
}

func (p *PrivilegedContainerExperiment) Framework() string {
	return string(categories.Mitre)
}

func (p *PrivilegedContainerExperiment) DependsOn() []string {
	return p.Metadata.DependsOn
}

func (p *PrivilegedContainerExperiment) Run(ctx context.Context, client *k8s.Client, experimentConfig *ExperimentConfig) error {
	var config PrivilegedContainerExperiment
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err := yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return err
	}

	if config.Parameters.Image == "" && len(config.Parameters.Command) == 0 {
		config.Parameters.Image = "alpine:latest"
		config.Parameters.Command = []string{
			"sh",
			"-c",
			"while true; do :; done",
		}
	}

	clientset := client.Clientset
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.Metadata.Name,
			Labels: map[string]string{
				"experiment": config.Metadata.Name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":        config.Metadata.Name,
					"experiment": config.Metadata.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"experiment": config.Metadata.Name,
						"app":        config.Metadata.Name,
					},
				},
				Spec: corev1.PodSpec{
					HostNetwork: config.Parameters.HostNetwork,
					HostPID:     config.Parameters.HostPid,
					Containers: []corev1.Container{
						{
							Name:            config.Metadata.Name,
							Image:           config.Parameters.Image,
							ImagePullPolicy: corev1.PullAlways,
							Command:         config.Parameters.Command,
						},
					},
				},
			},
		},
	}

	params := config.Parameters
	container := deployment.Spec.Template.Spec.Containers[0]
	securityContext := &corev1.SecurityContext{}
	if params.RunAsRoot {
		securityContext.RunAsUser = pointer.Int64(0)
	}

	if params.Privileged {
		securityContext.Privileged = pointer.Bool(true)
	}

	container.SecurityContext = securityContext
	deployment.Spec.Template.Spec.Containers[0] = container

	_, err = clientset.AppsV1().Deployments(config.Metadata.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	return err
}

func (p *PrivilegedContainerExperiment) Verify(ctx context.Context, client *k8s.Client, experimentConfig *ExperimentConfig) (*verifier.Outcome, error) {
	var config PrivilegedContainerExperiment
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err := yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return nil, err
	}

	clientset := client.Clientset
	deployment, err := clientset.AppsV1().Deployments(config.Metadata.Namespace).Get(ctx, config.Metadata.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	params := config.Parameters

	verifier := verifier.New(
		config.Metadata.Name,
		config.Description(),
		config.Framework(),
		config.Tactic(),
		config.Technique(),
	)

	if params.HostPid {
		verifier.Fail("HostPID")
		if deployment.Spec.Template.Spec.HostPID {
			verifier.Success("HostPID")
		}
	}

	if params.HostNetwork {
		verifier.Fail("HostNetwork")
		if deployment.Spec.Template.Spec.HostNetwork {
			verifier.Success("HostNetwork")
		}
	}

	// Find the container by name, as it may not be the first container in the list due to sidecar injection
	container, err := k8s.FindContainerByName(deployment.Spec.Template.Spec.Containers, config.Metadata.Name)
	if err != nil {
		return nil, err
	}

	if params.RunAsRoot {
		verifier.Fail("RunAsRoot")
		if container.SecurityContext != nil {
			if container.SecurityContext.RunAsUser != nil {
				if *container.SecurityContext.RunAsUser == 0 {
					verifier.Success("RunAsRoot")
				}
			}
		}
	}

	if params.Privileged {
		verifier.Fail("Privileged")
		if container.SecurityContext != nil {
			if container.SecurityContext.Privileged != nil {
				if *container.SecurityContext.Privileged {
					verifier.Success("Privileged")
				}
			}
		}
	}

	return verifier.GetOutcome(), nil
}

func (p *PrivilegedContainerExperiment) Cleanup(ctx context.Context, client *k8s.Client, experimentConfig *ExperimentConfig) error {
	var config PrivilegedContainerExperiment
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err := yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return err
	}
	clientset := client.Clientset
	return clientset.AppsV1().Deployments(config.Metadata.Namespace).Delete(ctx, config.Metadata.Name, metav1.DeleteOptions{})
}
