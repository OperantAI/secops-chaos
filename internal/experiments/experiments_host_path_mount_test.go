package experiments

import (
	corev1 "k8s.io/api/core/v1"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckVolumes(t *testing.T) {
	tests := []struct {
		name         string
		pod          corev1.Pod
		path         string
		expectResult bool
	}{
		{
			name: "Positive mount",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "workload-data",
						},
						{
							Name: "host-path",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "host-path-volume",
								},
							},
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: "Running",
				},
			},
			path:         "host-path-volume",
			expectResult: true,
		},
		{
			name: "Negative mount",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "workload-data",
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: "Running",
				},
			},
			path:         "host-path-volume",
			expectResult: false,
		},
		{
			name: "Zero mount",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{},
			},
			path:         "host-path-volume",
			expectResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := checkVolumes(test.pod, test.path)
			if test.expectResult {
				assert.True(t, result)
			} else {
				assert.False(t, result)
			}
		})
	}
}
