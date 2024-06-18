package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
)

func TestFindContainerByName(t *testing.T) {
	tests := []struct {
		name           string
		containers     []corev1.Container
		containerName  string
		expectedResult corev1.Container
		expectedError  error
	}{
		{
			name: "Container found in the list",
			containers: []corev1.Container{
				{Name: "container1"},
				{Name: "container2"},
			},
			containerName: "container2",
			expectedResult: corev1.Container{
				Name: "container2",
			},
			expectedError: nil,
		},
		{
			name: "Container not found in the list",
			containers: []corev1.Container{
				{Name: "container1"},
				{Name: "container2"},
			},
			containerName:  "nonexistent",
			expectedResult: corev1.Container{},
			expectedError:  ErrContainerNotFound,
		},
		{
			name:           "Empty container list",
			containers:     []corev1.Container{},
			containerName:  "container1",
			expectedResult: corev1.Container{},
			expectedError:  ErrContainerNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := Client{}
			result, err := client.FindContainerByName(test.containers, test.containerName)

			assert.Equal(t, test.expectedResult, result)
			assert.Equal(t, test.expectedError, err)
		})
	}
}
