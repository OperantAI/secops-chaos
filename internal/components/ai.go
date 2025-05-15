package components

import (
	"context"

	"github.com/operantai/woodpecker/internal/k8s"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
)

const (
	defaultSecopsChaosAIImage = "ghcr.io/operantai/woodpecker-ai:latest"
)

type AI struct{}

func (ai *AI) Type() string {
	return "woodpecker-ai"
}

func (ai *AI) Description() string {
	return "Enables running Experiments against AI providers"
}

func (ai *AI) Install(ctx context.Context, client *k8s.Client, config *Config) error {
	if config.Image == "" {
		config.Image = defaultSecopsChaosAIImage
	}

	err := checkForSecret(ctx, client, config)
	if err != nil {
		return err
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.Type,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": config.Type,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": config.Type,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            config.Type,
							Image:           config.Image,
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
												Name: config.Type,
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
	_, err = client.Clientset.AppsV1().Deployments(config.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.Type,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": config.Type,
			},
			Ports: []corev1.ServicePort{
				{
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
				},
			},
		},
	}
	_, err = client.Clientset.CoreV1().Services(config.Namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (ai *AI) Uninstall(ctx context.Context, client *k8s.Client, config *Config) error {
	err := client.Clientset.AppsV1().Deployments(config.Namespace).Delete(ctx, config.Type, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	err = client.Clientset.CoreV1().Services(config.Namespace).Delete(ctx, config.Type, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}
