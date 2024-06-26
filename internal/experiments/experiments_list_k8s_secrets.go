/*
Copyright 2023 Operant AI
*/
package experiments

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/operantai/secops-chaos/internal/executor"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/operantai/secops-chaos/internal/categories"
	"github.com/operantai/secops-chaos/internal/k8s"
	"github.com/operantai/secops-chaos/internal/verifier"
)

type ListK8sSecrets struct {
	Metadata   ExperimentMetadata
	Parameters K8sSecretsParameters
}

type K8sSecretsParameters struct {
	ExecutorConfig executor.RemoteExecuteAPI `yaml:"executorConfig"`
	Namespaces     []string                  `yaml:"namespaces"`
}

func (p *ListK8sSecrets) Name() string {
	return p.Metadata.Name
}

func (p *ListK8sSecrets) Type() string {
	return "list-kubernetes-secrets"
}

func (p *ListK8sSecrets) Description() string {
	return "List Kubernetes secrets in namespaces from within a container"
}

func (p *ListK8sSecrets) Technique() string {
	return categories.MITRE.Credentials.ListK8sSecrets.Technique
}

func (p *ListK8sSecrets) Tactic() string {
	return categories.MITRE.Credentials.ListK8sSecrets.Tactic
}

func (p *ListK8sSecrets) Framework() string {
	return string(categories.Mitre)
}

func (p *ListK8sSecrets) DependsOn() []string {
	return p.Metadata.DependsOn
}

func (p *ListK8sSecrets) Run(ctx context.Context, client *k8s.Client) error {
	executorConfig := executor.NewExecutorConfig(
		p.Metadata.Name,
		p.Metadata.Namespace,
		p.Parameters.ExecutorConfig.Image,
		p.Parameters.ExecutorConfig.ImageParameters,
		p.Parameters.ExecutorConfig.ServiceAccountName,
		p.Parameters.ExecutorConfig.Target.Port,
	)
	clusterrole := &v1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: p.Metadata.Name,
		},
		Rules: []v1.PolicyRule{
			{
				Verbs: []string{
					"list",
					"get",
				},
				Resources: []string{
					"secrets",
				},
				APIGroups: []string{
					"",
				},
			},
		},
	}
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.Parameters.ExecutorConfig.ServiceAccountName,
			Namespace: p.Metadata.Namespace,
		},
	}
	clusterRoleBinding := &v1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: p.Metadata.Name,
		},
		Subjects: []v1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      p.Parameters.ExecutorConfig.ServiceAccountName,
				Namespace: p.Metadata.Namespace,
				APIGroup:  "",
			},
		},
		RoleRef: v1.RoleRef{
			Kind:     "ClusterRole",
			Name:     p.Metadata.Name,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
	_, err := client.Clientset.RbacV1().ClusterRoles().Create(ctx, clusterrole, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	_, err = client.Clientset.CoreV1().ServiceAccounts(p.Metadata.Namespace).Create(ctx, serviceAccount, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	_, err = client.Clientset.RbacV1().ClusterRoleBindings().Create(ctx, clusterRoleBinding, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	err = executorConfig.Deploy(ctx, client.Clientset)
	if err != nil {
		return err
	}

	return nil

}

func (p *ListK8sSecrets) Verify(ctx context.Context, client *k8s.Client) (*verifier.Outcome, error) {
	v := verifier.New(
		p.Metadata.Name,
		p.Description(),
		p.Framework(),
		p.Tactic(),
		p.Technique(),
	)

	pf := client.NewPortForwarder(ctx)
	defer pf.Stop()
	forwardedPort, err := pf.Forward(
		p.Metadata.Namespace,
		fmt.Sprintf("app=%s", p.Metadata.Name),
		int(p.Parameters.ExecutorConfig.Target.Port),
	)
	if err != nil {
		return nil, err
	}

	path := p.Parameters.ExecutorConfig.Target.Path
	for _, namespace := range p.Parameters.Namespaces {
		requestUrl := url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%d", pf.Addr(), int32(forwardedPort.Local)),
			Path:   fmt.Sprintf("%s/%s", path, namespace),
		}
		response, err := http.Get(requestUrl.String())
		if err != nil {
			return nil, err
		}
		if response.StatusCode == http.StatusOK {
			v.Success(namespace)
		} else {
			v.Fail(namespace)
		}
		defer response.Body.Close()
	}
	return v.GetOutcome(), nil
}

func (p *ListK8sSecrets) Cleanup(ctx context.Context, client *k8s.Client) error {
	clientset := client.Clientset
	executorConfig := executor.NewExecutorConfig(
		p.Metadata.Name,
		p.Metadata.Namespace,
		p.Parameters.ExecutorConfig.Image,
		p.Parameters.ExecutorConfig.ImageParameters,
		p.Parameters.ExecutorConfig.ServiceAccountName,
		p.Parameters.ExecutorConfig.Target.Port,
	)

	err := executorConfig.Cleanup(ctx, clientset)
	if err != nil {
		return err
	}
	err = client.Clientset.RbacV1().ClusterRoleBindings().Delete(ctx, p.Metadata.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	err = client.Clientset.RbacV1().ClusterRoles().Delete(ctx, p.Metadata.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	err = client.Clientset.CoreV1().ServiceAccounts(p.Metadata.Namespace).Delete(ctx, p.Parameters.ExecutorConfig.ServiceAccountName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}
