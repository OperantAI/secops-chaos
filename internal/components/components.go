package components

import (
	"context"
	"os"

	"github.com/operantai/woodpecker/internal/output"

	"gopkg.in/yaml.v3"
)

var registry = map[string]Component{
	"woodpecker-ai": &AI{},
}

type Installer struct {
	ctx context.Context
}

type Component interface {
	// Type returns the type of component
	Type() string
	// Description describes what the component does in a brief sentence
	Description() string
	// Install installs the Component
	Install(ctx context.Context, config *Config) error
	// Uninstall uninstalls the Component
	Uninstall(ctx context.Context, config *Config) error
}

type Config struct {
	Type       string   `yaml:"type"`
	Namespace  string   `yaml:"namespace"`
	Image      string   `yaml:"image"`
	SecretName string   `yaml:"secretName"`
	SecretEnvs []string `yaml:"secretEnvs"`
}

func New(ctx context.Context) *Installer {
	return &Installer{
		ctx: ctx,
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

		output.WriteInfo("Adding component %s", config.Type)
		component := registry[config.Type]
		if err := component.Install(i.ctx, &config); err != nil {
			output.WriteFatal("Could not install component %s: %s", config.Type, err.Error())
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
		if err := component.Uninstall(i.ctx, &config); err != nil {
			output.WriteFatal("Could not uninstall component %s: %s", config.Type, err.Error())
		}
	}
	return nil
}
