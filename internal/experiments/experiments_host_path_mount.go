/*
Copyright 2023 Operant AI
*/
package experiments

import (
	"context"
	"fmt"

	"github.com/operantai/secops-chaos/internal/categories"
	"github.com/operantai/secops-chaos/internal/k8s"
	"github.com/operantai/secops-chaos/internal/verifier"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

type HostPathMountExperiment struct {
	Metadata   ExperimentMetadata
	Parameters HostPathMountParameters
}

type HostPathMountParameters struct {
	HostPath HostPath `yaml:"hostPath"`
}

type HostPath struct {
	Path string `yaml:"path"`
}

func (p *HostPathMountExperiment) Name() string {
	return p.Name()
}

func (p *HostPathMountExperiment) Type() string {
	return "host-path-mount"
}

func (p *HostPathMountExperiment) Description() string {
	return "Mount a sensitive host filesystem path into a container"
}

func (p *HostPathMountExperiment) Technique() string {
	return categories.MITRE.PrivilegeEscalation.HostPathMount.Technique
}

func (p *HostPathMountExperiment) Tactic() string {
	return categories.MITRE.PrivilegeEscalation.HostPathMount.Tactic
}

func (p *HostPathMountExperiment) Framework() string {
	return string(categories.Mitre)
}

func (p *HostPathMountExperiment) DependsOn() []string {
	return p.Metadata.DependsOn
}

func (p *HostPathMountExperiment) Run(ctx context.Context, client *k8s.Client) error {
	clientset := client.Clientset
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: p.Metadata.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": p.Metadata.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"experiment": p.Metadata.Name,
						"app":        p.Metadata.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            p.Metadata.Name,
							Image:           "alpine:latest",
							ImagePullPolicy: corev1.PullAlways,
							Command: []string{
								"sh",
								"-c",
								"while true; do :; done",
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 4000,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "hostpath-volume",
									MountPath: "/tmp",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "hostpath-volume",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: p.Parameters.HostPath.Path,
								},
							},
						},
					},
				},
			},
		},
	}
	_, err := clientset.AppsV1().Deployments(p.Metadata.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	return err
}

func (p *HostPathMountExperiment) Verify(ctx context.Context, client *k8s.Client) (*verifier.Outcome, error) {
	v := verifier.New(
		p.Metadata.Name,
		p.Description(),
		p.Framework(),
		p.Tactic(),
		p.Technique(),
	)
	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", p.Metadata.Name),
	}
	clientset := client.Clientset
	pods, err := clientset.CoreV1().Pods(p.Metadata.Namespace).List(ctx, listOptions)
	if err != nil {
		return nil, err
	}
	if len(pods.Items) == 1 {
		if checkVolumes(pods.Items[0], p.Parameters.HostPath.Path) {
			v.Success("")
		} else {
			v.Fail("")
		}
	}
	return v.GetOutcome(), nil
}

func checkVolumes(pod corev1.Pod, volumePath string) bool {
	if pod.Status.Phase == "Running" {
		volumes := pod.Spec.Volumes
		for _, v := range volumes {
			if v.HostPath != nil && v.HostPath.Path == volumePath {
				return true
			}
		}
	}
	return false
}

func (p *HostPathMountExperiment) Cleanup(ctx context.Context, client *k8s.Client) error {
	clientset := client.Clientset
	return clientset.AppsV1().Deployments(p.Metadata.Namespace).Delete(ctx, p.Metadata.Name, metav1.DeleteOptions{})
}
