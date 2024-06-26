/*
Copyright 2023 Operant AI
*/
package experiments

import (
	"context"

	"github.com/operantai/secops-chaos/internal/categories"
	"github.com/operantai/secops-chaos/internal/k8s"
	"github.com/operantai/secops-chaos/internal/verifier"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

type ClusterAdminBindingExperiment struct {
	Metadata   ExperimentMetadata
	Parameters ClusterAdminBinding
}

type ClusterAdminBinding struct{}

func (p *ClusterAdminBindingExperiment) Name() string {
	return p.Metadata.Name
}

func (p *ClusterAdminBindingExperiment) Type() string {
	return "cluster-admin-binding"
}

func (p *ClusterAdminBindingExperiment) Description() string {
	return "Create a container with the cluster-admin role binding attached"
}

func (p *ClusterAdminBindingExperiment) Technique() string {
	return categories.MITRE.PrivilegeEscalation.ClusterAdminBinding.Technique
}

func (p *ClusterAdminBindingExperiment) Tactic() string {
	return categories.MITRE.PrivilegeEscalation.ClusterAdminBinding.Tactic
}

func (p *ClusterAdminBindingExperiment) Framework() string {
	return string(categories.Mitre)
}

func (p *ClusterAdminBindingExperiment) DependsOn() []string {
	return p.Metadata.DependsOn
}

func (p *ClusterAdminBindingExperiment) Run(ctx context.Context, client *k8s.Client) error {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.Metadata.Name,
			Namespace: p.Metadata.Namespace,
			Labels: map[string]string{
				"experiment": p.Metadata.Name,
			},
		},
	}

	clientset := client.Clientset
	_, err := clientset.CoreV1().ServiceAccounts(p.Metadata.Namespace).Create(ctx, sa, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: p.Metadata.Name,
			Labels: map[string]string{
				"experiment": p.Metadata.Name,
			},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      p.Metadata.Name,
				Namespace: p.Metadata.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	_, err = clientset.RbacV1().ClusterRoleBindings().Create(ctx, clusterRoleBinding, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: p.Metadata.Name,
			Labels: map[string]string{
				"experiment": p.Metadata.Name,
			},
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
					ServiceAccountName: sa.ObjectMeta.Name,
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
						},
					},
				},
			},
		},
	}
	_, err = clientset.AppsV1().Deployments(p.Metadata.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	return err
}

func (p *ClusterAdminBindingExperiment) Verify(ctx context.Context, client *k8s.Client) (*verifier.Outcome, error) {
	v := verifier.New(
		p.Metadata.Name,
		p.Description(),
		p.Framework(),
		p.Tactic(),
		p.Technique(),
	)

	listOptions := metav1.ListOptions{
		LabelSelector: "experiment=" + p.Metadata.Name,
	}

	clientset := client.Clientset
	pods, err := clientset.CoreV1().Pods(p.Metadata.Namespace).List(ctx, listOptions)
	if err != nil {
		return nil, err
	}

	// Check if the pod is running and has the service account attached
	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodRunning {
			if pod.Spec.ServiceAccountName == p.Metadata.Name {
				v.Success(p.Metadata.Type)
			} else {
				v.Fail(p.Metadata.Type)
			}
		} else {
			v.Fail(p.Metadata.Type)
		}
	}

	return v.GetOutcome(), nil
}

func (p *ClusterAdminBindingExperiment) Cleanup(ctx context.Context, client *k8s.Client) error {
	clientset := client.Clientset
	err := clientset.AppsV1().Deployments(p.Metadata.Namespace).Delete(ctx, p.Metadata.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	err = clientset.RbacV1().ClusterRoleBindings().Delete(ctx, p.Metadata.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	err = clientset.CoreV1().ServiceAccounts(p.Metadata.Namespace).Delete(ctx, p.Metadata.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}
