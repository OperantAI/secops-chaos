package k8s

import (
	"errors"

	corev1 "k8s.io/api/core/v1"
)

// ErrContainerNotFound is returned when a container is not found
var ErrContainerNotFound = errors.New("container not found")

// FindContainerByName returns a container by name from a list of containers
func FindContainerByName(containers []corev1.Container, containerName string) (corev1.Container, error) {
	for _, container := range containers {
		if container.Name == containerName {
			return container, nil
		}
	}
	return corev1.Container{}, ErrContainerNotFound
}
