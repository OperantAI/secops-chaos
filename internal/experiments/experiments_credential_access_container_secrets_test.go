package experiments

import (
	corev1 "k8s.io/api/core/v1"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckPodForSecrets(t *testing.T) {
	tests := []struct {
		name         string
		pod          corev1.Pod
		env          []ContainerSecretsEnv
		expectResult bool
	}{
		{
			name: "Positive Pod secrets test",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Env: []corev1.EnvVar{
								{
									Name:  "AWS_SECRET",
									Value: "AWS_SECRET_VALUE",
								},
							},
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: "Running",
				},
			},
			env: []ContainerSecretsEnv{
				{
					EnvKey:   "AWS",
					EnvValue: "AWS_SECRET_VALUE",
				},
			},
			expectResult: true,
		},
		{
			name: "Negative Pod secrets test",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Env: []corev1.EnvVar{
								{
									Name:  "EnV",
									Value: "Dev",
								},
							},
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: "Running",
				},
			},
			env: []ContainerSecretsEnv{
				{
					EnvKey:   "AWS",
					EnvValue: "AWS_SECRET_VALUE",
				},
			},
			expectResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := checkPodForSecrets(test.pod, test.env)
			if test.expectResult {
				assert.True(t, result)
			} else {
				assert.False(t, result)
			}
		})
	}
}

func TestCheckConfigMapForSecrets(t *testing.T) {
	tests := []struct {
		name         string
		cm           corev1.ConfigMap
		env          []ContainerSecretsEnv
		expectResult bool
	}{
		{
			name: "Positive CM secrets test",
			cm: corev1.ConfigMap{
				Data: map[string]string{"AWS_ACCESS_KEY": "AWS_ACCESS_KEY_VALUE"},
			},
			env: []ContainerSecretsEnv{
				{
					EnvKey:   "AWS_ACCESS_KEY",
					EnvValue: "AWS_SECRET_VALUE",
				},
			},
			expectResult: true,
		},
		{
			name: "Negative CM secrets test",
			cm: corev1.ConfigMap{
				Data: map[string]string{"Env": "Prod"},
			},
			env: []ContainerSecretsEnv{
				{
					EnvKey:   "AWS_ACCESS_KEY",
					EnvValue: "AWS_SECRET_VALUE",
				},
			},
			expectResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := checkConfigMapForSecrets(test.cm, test.env)
			if test.expectResult {
				assert.True(t, result)
			} else {
				assert.False(t, result)
			}
		})
	}
}
