package k8s

import (
	"context"
	"errors"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ErrContainerNotFound is returned when a container is not found
var ErrContainerNotFound = errors.New("container not found")

// FindContainerByName returns a container by name from a list of containers
func (c *Client) FindContainerByName(containers []corev1.Container, containerName string) (corev1.Container, error) {
	for _, container := range containers {
		if container.Name == containerName {
			return container, nil
		}
	}
	return corev1.Container{}, ErrContainerNotFound
}

// GetDeploymentsPods gets the pods belonging to the provided deployment
func (c *Client) GetDeploymentsPods(ctx context.Context, namespace string, deployment *appsv1.Deployment) ([]corev1.Pod, error) {
	rsList, err := c.Clientset.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(deployment.Spec.Selector),
	})
	if err != nil {
		return nil, err
	}
	var podsList []corev1.Pod
	for _, rs := range rsList.Items {
		pods, err := c.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(rs.Spec.Selector),
		})
		if err != nil {
			return nil, err
		}
		for _, pod := range pods.Items {
			podsList = append(podsList, pod)
		}
	}
	return podsList, nil
}
