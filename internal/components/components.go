package components

import (
	"context"
	"fmt"
	"os"

	"github.com/operantai/secops-chaos/internal/k8s"
	"github.com/operantai/secops-chaos/internal/output"

	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
)

const (
	defaultSecopsChaosAIImage = "ghcr.io/operantai/secops-chaos-ai:latest"
)

type Components struct {
	ctx context.Context
	k8s *k8s.Client
}

type Component struct {
	Type       string `yaml:"type"`
	Namespace  string `yaml:"namespace"`
	Image      string `yaml:"image"`
	SecretName string `yaml:"secretName"`
}

func New(ctx context.Context) *Components {
	k8sClient, err := k8s.NewClient()
	if err != nil {
		output.WriteFatal("Error creating Kubernetes Client: %v", err)
	}
	return &Components{
		ctx: ctx,
		k8s: k8sClient,
	}
}

func (c *Components) Add(files []string) error {
	for _, file := range files {
		contents, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		var component Component
		if err := yaml.Unmarshal(contents, &component); err != nil {
			return err
		}

		output.WriteInfo("Adding component %s to Cluster", component.Type)
		switch component.Type {
		case "secops-chaos-ai":
			return c.installSecOpsChaosAI(&component)
		default:
			return fmt.Errorf("Unknown component %s", component.Type)
		}
	}
	return nil
}

func (c *Components) Remove(files []string) error {
	for _, file := range files {
		contents, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		var component Component
		if err := yaml.Unmarshal(contents, &component); err != nil {
			return err
		}

		output.WriteInfo("Removing component %s from Cluster", component.Type)
		switch component.Type {
		case "secops-chaos-ai":
			return c.uninstallSecOpsChaosAI(&component)
		default:
			return fmt.Errorf("Unknown component %s", component.Type)
		}
	}
	return nil
}

func (c *Components) installSecOpsChaosAI(component *Component) error {
	if component.Image == "" {
		component.Image = defaultSecopsChaosAIImage
	}

	err := c.checkForSecret(component)
	if err != nil {
		return err
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: component.Type,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": component.Type,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": component.Type,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            component.Type,
							Image:           component.Image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name: "OPENAI_KEY",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: component.Type,
											},
											Key: "OPENAI_KEY",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	_, err = c.k8s.Clientset.AppsV1().Deployments(component.Namespace).Create(c.ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: component.Type,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": component.Type,
			},
			Ports: []corev1.ServicePort{
				{
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
				},
			},
		},
	}
	_, err = c.k8s.Clientset.CoreV1().Services(component.Namespace).Create(c.ctx, service, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *Components) uninstallSecOpsChaosAI(component *Component) error {
	err := c.k8s.Clientset.AppsV1().Deployments(component.Namespace).Delete(c.ctx, component.Type, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	err = c.k8s.Clientset.CoreV1().Services(component.Namespace).Delete(c.ctx, component.Type, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *Components) checkForSecret(component *Component) error {
	_, err := c.k8s.Clientset.CoreV1().Secrets(component.Namespace).Get(c.ctx, component.SecretName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Could not find secret: %w", err)
	}
	return nil
}
