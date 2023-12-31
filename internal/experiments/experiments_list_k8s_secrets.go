/*
Copyright 2023 Operant AI
*/
package experiments

import (
	"context"
	"fmt"
	"github.com/operantai/secops-chaos/internal/executor"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"net/url"

	"github.com/operantai/secops-chaos/internal/categories"
	"github.com/operantai/secops-chaos/internal/k8s"
	"github.com/operantai/secops-chaos/internal/verifier"
	"gopkg.in/yaml.v3"
)

type ListK8sSecretsConfig struct {
	Metadata   ExperimentMetadata   `yaml:"metadata"`
	Parameters K8sSecretsParameters `yaml:"parameters"`
}

type K8sSecretsParameters struct {
	ExecutorConfig executor.RemoteExecuteAPI `yaml:"executor_config"`
	Namespaces     []string                  `yaml:"namespaces"`
}

func (p *ListK8sSecretsConfig) Type() string {
	return "list_kubernetes_secrets"
}

func (p *ListK8sSecretsConfig) Description() string {
	return "This experiment attempts to list Kubernetes secrets in different namespaces from within a container"
}

func (p *ListK8sSecretsConfig) Technique() string {
	return categories.MITRE.Credentials.ListK8sSecrets.Technique
}

func (p *ListK8sSecretsConfig) Tactic() string {
	return categories.MITRE.Credentials.ListK8sSecrets.Tactic
}

func (p *ListK8sSecretsConfig) Framework() string {
	return string(categories.Mitre)
}

func (p *ListK8sSecretsConfig) Run(ctx context.Context, client *k8s.Client, experimentConfig *ExperimentConfig) error {
	var config ListK8sSecretsConfig
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err := yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return err
	}

	executorConfig := executor.NewExecutorConfig(
		config.Metadata.Name,
		config.Metadata.Namespace,
		config.Parameters.ExecutorConfig.Image,
		config.Parameters.ExecutorConfig.ImageParameters,
		config.Parameters.ExecutorConfig.ServiceAccountName,
		config.Parameters.ExecutorConfig.Target.Port,
	)
	clusterrole := &v1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.Metadata.Name,
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
			Name:      config.Parameters.ExecutorConfig.ServiceAccountName,
			Namespace: config.Metadata.Namespace,
		},
	}
	clusterRoleBinding := &v1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.Metadata.Name,
		},
		Subjects: []v1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      config.Parameters.ExecutorConfig.ServiceAccountName,
				Namespace: config.Metadata.Namespace,
				APIGroup:  "",
			},
		},
		RoleRef: v1.RoleRef{
			Kind:     "ClusterRole",
			Name:     config.Metadata.Name,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
	_, err = client.Clientset.RbacV1().ClusterRoles().Create(ctx, clusterrole, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	_, err = client.Clientset.CoreV1().ServiceAccounts(config.Metadata.Namespace).Create(ctx, serviceAccount, metav1.CreateOptions{})
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

func (p *ListK8sSecretsConfig) Verify(ctx context.Context, client *k8s.Client, experimentConfig *ExperimentConfig) (*verifier.Outcome, error) {
	var config ListK8sSecretsConfig
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
	executorConfig := executor.NewExecutorConfig(
		config.Metadata.Name,
		config.Metadata.Namespace,
		config.Parameters.ExecutorConfig.Image,
		config.Parameters.ExecutorConfig.ImageParameters,
		config.Parameters.ExecutorConfig.ServiceAccountName,
		config.Parameters.ExecutorConfig.Target.Port,
	)

	forwarder, err := executorConfig.OpenLocalPort(ctx, client)
	if err != nil {
		return nil, err
	}

	go func() {
		err = forwarder.ForwardPorts()
		if err != nil {
			executorConfig.ErrCh <- fmt.Errorf("Local Port Failed to open: %w", err)
		}
	}()

	// Waits until local port is ready or open errors
	select {
	case <-executorConfig.ReadyCh:
		break
	case err := <-executorConfig.ErrCh:
		return nil, err
	}

	path := config.Parameters.ExecutorConfig.Target.Path
	ports, err := forwarder.GetPorts()
	if err != nil {
		return nil, err
	}
	for _, namespace := range config.Parameters.Namespaces {
		requestUrl := url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%d", executorConfig.Addr, int32(ports[0].Local)),
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

func (p *ListK8sSecretsConfig) Cleanup(ctx context.Context, client *k8s.Client, experimentConfig *ExperimentConfig) error {
	clientset := client.Clientset
	var config ListK8sSecretsConfig
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err := yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return err
	}

	executorConfig := executor.NewExecutorConfig(
		config.Metadata.Name,
		config.Metadata.Namespace,
		config.Parameters.ExecutorConfig.Image,
		config.Parameters.ExecutorConfig.ImageParameters,
		config.Parameters.ExecutorConfig.ServiceAccountName,
		config.Parameters.ExecutorConfig.Target.Port,
	)

	err = executorConfig.Cleanup(ctx, clientset)
	if err != nil {
		return err
	}
	err = client.Clientset.RbacV1().ClusterRoleBindings().Delete(ctx, config.Metadata.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	err = client.Clientset.RbacV1().ClusterRoles().Delete(ctx, config.Metadata.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	err = client.Clientset.CoreV1().ServiceAccounts(config.Metadata.Namespace).Delete(ctx, config.Parameters.ExecutorConfig.ServiceAccountName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}
