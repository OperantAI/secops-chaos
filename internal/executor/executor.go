package executor

import (
	"context"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/pointer"
)

const (
	addr string = "127.0.0.1"
)

type RemoteExecutorConfig struct {
	Name       string
	Namespace  string
	Image      string
	Parameters RemoteExecutor
}

type RemoteExecutor struct {
	ServiceAccountName string
	TargetPort         int32
	ImageParameters    []string
}
type RemoteExecuteAPI struct {
	Image              string   `yaml:"image"`
	ImageParameters    []string `yaml:"imageParameters"`
	ServiceAccountName string   `yaml:"serviceAccountName"`
	Target             Target   `yaml:"target"`
}
type Target struct {
	Port int32  `yaml:"targetPort"`
	Path string `yaml:"path"`
}

// Executor configurations are meant to be used to execute remote commands on a pod in a cluster.
func NewExecutorConfig(name, namespace, image string, imageParameters []string, serviceAccountName string, targetPort int32) *RemoteExecutorConfig {
	return &RemoteExecutorConfig{
		Name:      name,
		Namespace: namespace,
		Image:     image,
		Parameters: RemoteExecutor{
			ServiceAccountName: serviceAccountName,
			TargetPort:         targetPort,
			ImageParameters:    imageParameters,
		},
	}
}

func (r *RemoteExecutorConfig) Deploy(ctx context.Context, client *kubernetes.Clientset) error {
	envVar := prepareImageParameters(r.Parameters.ImageParameters)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: r.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": r.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"experiment": r.Name,
						"app":        r.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            r.Name,
							Image:           r.Image,
							ImagePullPolicy: corev1.PullAlways,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: r.Parameters.TargetPort,
								},
							},
							Env: envVar,
						},
					},
				},
			},
		},
	}
	params := r.Parameters
	if params.ServiceAccountName != "" {
		deployment.Spec.Template.Spec.ServiceAccountName = params.ServiceAccountName
	}

	_, err := client.AppsV1().Deployments(r.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: r.Name,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": r.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Port: r.Parameters.TargetPort,
				},
			},
		},
	}

	_, err = client.CoreV1().Services(r.Namespace).Create(ctx, service, metav1.CreateOptions{})
	return err
}

func (r *RemoteExecutorConfig) Cleanup(ctx context.Context, client *kubernetes.Clientset) error {
	err := client.AppsV1().Deployments(r.Namespace).Delete(ctx, r.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return client.CoreV1().Services(r.Namespace).Delete(ctx, r.Name, metav1.DeleteOptions{})
}

func prepareImageParameters(imageParameters []string) []corev1.EnvVar {
	var envVar []corev1.EnvVar
	for _, param := range imageParameters {
		parts := strings.Split(param, "=")
		envVar = append(envVar, corev1.EnvVar{
			Name:  parts[0],
			Value: parts[1],
		})
	}

	return envVar
}
