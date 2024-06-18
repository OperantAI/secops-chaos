package components

import (
	"context"
	"fmt"
	"os"

	"github.com/operantai/secops-chaos/internal/k8s"
	"github.com/operantai/secops-chaos/internal/output"

	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var registry = map[string]Component{
	"secops-chaos-ai": &AI{},
}

type Installer struct {
	ctx context.Context
	k8s *k8s.Client
}

type Component interface {
	// Type returns the type of component
	Type() string
	// Description describes what the component does in a brief sentence
	Description() string
	// Install installs the Component to the Kubernetes Cluster
	Install(ctx context.Context, client *k8s.Client, config *Config) error
	// Uninstall uninstalls the Compoent from the Kubernetes Cluster
	Uninstall(ctx context.Context, client *k8s.Client, config *Config) error
}

type Config struct {
	Type       string `yaml:"type"`
	Namespace  string `yaml:"namespace"`
	Image      string `yaml:"image"`
	SecretName string `yaml:"secretName"`
}

func New(ctx context.Context) *Installer {
	k8sClient, err := k8s.NewClient()
	if err != nil {
		output.WriteFatal("Error creating Kubernetes Client: %v", err)
	}
	return &Installer{
		ctx: ctx,
		k8s: k8sClient,
	}
}

func ListComponents() map[string]string {
	components := make(map[string]string)
	for _, component := range registry {
		components[component.Type()] = component.Description()
	}
	return components
}

func (i *Installer) Add(files []string) error {
	for _, file := range files {
		contents, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		var config Config
		if err := yaml.Unmarshal(contents, &config); err != nil {
			return err
		}

		output.WriteInfo("Adding component %s to Cluster", config.Type)
		if err := i.checkNamespaceExists(&config); err != nil {
			return err
		}
		component := registry[config.Type]
		if err := component.Install(i.ctx, i.k8s, &config); err != nil {
			output.WriteFatal("Could not install component: %s", config.Type)
		}
	}
	return nil
}

func (i *Installer) Remove(files []string) error {
	for _, file := range files {
		contents, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		var config Config
		if err := yaml.Unmarshal(contents, &config); err != nil {
			return err
		}

		output.WriteInfo("Removing component %s from Cluster", config.Type)
		component := registry[config.Type]
		if err := component.Uninstall(i.ctx, i.k8s, &config); err != nil {
			output.WriteFatal("Could not uninstall component: %s", config.Type)
		}
	}
	return nil
}

func (c *Installer) checkNamespaceExists(component *Config) error {
	_, err := c.k8s.Clientset.CoreV1().Namespaces().Get(c.ctx, component.Namespace, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Could not find namespace: %w", err)
	}
	return nil
}

func checkForSecret(ctx context.Context, client *k8s.Client, component *Config) error {
	_, err := client.Clientset.CoreV1().Secrets(component.Namespace).Get(ctx, component.SecretName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Could not find secret: %w", err)
	}
	return nil
}
