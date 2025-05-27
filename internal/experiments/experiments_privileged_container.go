/*
Copyright 2023 Operant AI
*/
package experiments

import (
	"context"

	"gopkg.in/yaml.v3"

	"github.com/operantai/woodpecker/internal/categories"
	"github.com/operantai/woodpecker/internal/k8s"
	"github.com/operantai/woodpecker/internal/verifier"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/utils/pointer"
)

type PrivilegedContainerExperimentConfig struct {
	Metadata   ExperimentMetadata  `yaml:"metadata"`
	Parameters PrivilegedContainer `yaml:"parameters"`
}

// PrivilegedContainer is an experiment that creates a deployment with a privileged container
type PrivilegedContainer struct {
	Experiment struct {
		Image       string   `yaml:"image"`
		Command     []string `yaml:"command"`
		Privileged  bool     `yaml:"privileged"`
		HostPid     bool     `yaml:"hostPid"`
		HostNetwork bool     `yaml:"hostNetwork"`
		RunAsRoot   bool     `yaml:"runAsRoot"`
	} `yaml:"experiment"`
	Verifier struct {
		DeployedSuccessfully bool     `yaml:"deployed"`
		Command              []string `yaml:"command"`
	} `yaml:"verifier"`
}

func (p *PrivilegedContainerExperimentConfig) Type() string {
	return "privileged-container"
}

func (p *PrivilegedContainerExperimentConfig) Description() string {
	return "Run a privileged container in a namespace"
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

func (p *PrivilegedContainerExperimentConfig) Run(ctx context.Context, experimentConfig *ExperimentConfig) error {
	client, err := k8s.NewClient()
	if err != nil {
		return err
	}
	var config PrivilegedContainerExperimentConfig
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err = yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return err
	}

	params := config.Parameters.Experiment
	if params.Image == "" && len(params.Command) == 0 {
		params.Image = "alpine:latest"
		params.Command = []string{
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
					HostNetwork: params.HostNetwork,
					HostPID:     params.HostPid,
					Containers: []corev1.Container{
						{
							Name:            config.Metadata.Name,
							Image:           params.Image,
							ImagePullPolicy: corev1.PullAlways,
							Command:         params.Command,
						},
					},
				},
			},
		},
	}

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

func (p *PrivilegedContainerExperimentConfig) Verify(ctx context.Context, experimentConfig *ExperimentConfig) (*verifier.Outcome, error) {
	client, err := k8s.NewClient()
	if err != nil {
		return nil, err
	}
	var config PrivilegedContainerExperimentConfig
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err = yaml.Unmarshal(yamlObj, &config)
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

	// Find the container by name, as it may not be the first container in the list due to sidecar injection
	container, err := client.FindContainerByName(deployment.Spec.Template.Spec.Containers, config.Metadata.Name)
	if err != nil {
		return nil, err
	}

	if config.Parameters.Verifier.DeployedSuccessfully {
		verifier.Success("Deployed")
		if params.Experiment.HostPid {
			if !deployment.Spec.Template.Spec.HostPID {
				verifier.Fail("Deployed")
			}
		}

		if params.Experiment.HostNetwork {
			if !deployment.Spec.Template.Spec.HostNetwork {
				verifier.Fail("Deployed")
			}
		}

		if params.Experiment.RunAsRoot {
			if container.SecurityContext != nil {
				if container.SecurityContext.RunAsUser != nil {
					if *container.SecurityContext.RunAsUser != 0 {
						verifier.Fail("Deployed")
					}
				}
			}
		}

		if params.Experiment.Privileged {
			if container.SecurityContext != nil {
				if container.SecurityContext.Privileged != nil {
					if !*container.SecurityContext.Privileged {
						verifier.Fail("Deployed")
					}
				}
			}
		}
	}

	if len(params.Verifier.Command) > 0 {
		verifier.Success("Command")
		pods, err := client.GetDeploymentsPods(ctx, config.Metadata.Namespace, deployment)
		if err != nil {
			return nil, err
		}
		for _, pod := range pods {
			_, _, err := client.ExecuteRemoteCommand(ctx, config.Metadata.Namespace, pod.Name, container.Name, config.Parameters.Verifier.Command)
			if err != nil {
				verifier.Fail("Command")
			}
		}
	}

	return verifier.GetOutcome(), nil
}

func (p *PrivilegedContainerExperimentConfig) Cleanup(ctx context.Context, experimentConfig *ExperimentConfig) error {
	client, err := k8s.NewClient()
	if err != nil {
		return err
	}
	var config PrivilegedContainerExperimentConfig
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err = yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return err
	}
	clientset := client.Clientset
	return clientset.AppsV1().Deployments(config.Metadata.Namespace).Delete(ctx, config.Metadata.Name, metav1.DeleteOptions{})
}
