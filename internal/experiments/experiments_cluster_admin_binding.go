/*
Copyright 2023 Operant AI
*/
package experiments

import (
	"context"

	"github.com/operantai/secops-chaos/internal/categories"
	"github.com/operantai/secops-chaos/internal/k8s"
	"github.com/operantai/secops-chaos/internal/verifier"

	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

type ClusterAdminBindingExperiment struct {
	*ExperimentConfig
}

type ClusterAdminBinding struct{}

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

func (p *ClusterAdminBindingExperiment) Run(ctx context.Context, client *k8s.Client, experimentConfig *ExperimentConfig) error {
	var config ClusterAdminBindingExperiment
	yamlObj, err := yaml.Marshal(experimentConfig)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return err
	}

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.Metadata.Name,
			Namespace: config.Metadata.Namespace,
			Labels: map[string]string{
				"experiment": config.Metadata.Name,
			},
		},
	}

	clientset := client.Clientset
	_, err = clientset.CoreV1().ServiceAccounts(config.Metadata.Namespace).Create(ctx, sa, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.Metadata.Name,
			Labels: map[string]string{
				"experiment": config.Metadata.Name,
			},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      config.Metadata.Name,
				Namespace: config.Metadata.Namespace,
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
			Name: config.Metadata.Name,
			Labels: map[string]string{
				"experiment": config.Metadata.Name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": config.Metadata.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"experiment": config.Metadata.Name,
						"app":        config.Metadata.Name,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: sa.ObjectMeta.Name,
					Containers: []corev1.Container{
						{
							Name:            config.Metadata.Name,
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
	_, err = clientset.AppsV1().Deployments(config.Metadata.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	return err
}

func (p *ClusterAdminBindingExperiment) Verify(ctx context.Context, client *k8s.Client, experimentConfig *ExperimentConfig) (*verifier.Outcome, error) {
	var config ClusterAdminBindingExperiment
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err := yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return nil, err
	}

	v := verifier.New(
		config.Metadata.Name,
		config.Description(),
		config.Framework(),
		config.Tactic(),
		config.Technique(),
	)

	listOptions := metav1.ListOptions{
		LabelSelector: "experiment=" + config.Metadata.Name,
	}

	clientset := client.Clientset
	pods, err := clientset.CoreV1().Pods(config.Metadata.Namespace).List(ctx, listOptions)
	if err != nil {
		return nil, err
	}

	// Check if the pod is running and has the service account attached
	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodRunning {
			if pod.Spec.ServiceAccountName == config.Metadata.Name {
				v.Success(config.Metadata.Type)
			} else {
				v.Fail(config.Metadata.Type)
			}
		} else {
			v.Fail(config.Metadata.Type)
		}
	}

	return v.GetOutcome(), nil
}

func (p *ClusterAdminBindingExperiment) Cleanup(ctx context.Context, client *k8s.Client, experimentConfig *ExperimentConfig) error {
	var config ClusterAdminBindingExperiment
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err := yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return err
	}

	clientset := client.Clientset
	err = clientset.AppsV1().Deployments(config.Metadata.Namespace).Delete(ctx, config.Metadata.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	err = clientset.RbacV1().ClusterRoleBindings().Delete(ctx, config.Metadata.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	err = clientset.CoreV1().ServiceAccounts(config.Metadata.Namespace).Delete(ctx, config.Metadata.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}
