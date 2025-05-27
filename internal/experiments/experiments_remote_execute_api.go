package experiments

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/operantai/woodpecker/internal/categories"
	"github.com/operantai/woodpecker/internal/executor"
	"github.com/operantai/woodpecker/internal/k8s"
	"github.com/operantai/woodpecker/internal/verifier"
	"gopkg.in/yaml.v3"
)

// RemoteExecuteAPI is an experiment that uses the remote executor to check a remote output
// The image must be created independently -- the current default is `alconen/egress_server`, which runs a simple web app on port 4000 that checks http connectivity to
// a few domains ("https://google.com", "https://linkedin.com", "https://openai.com/") and responds with a success based on the success of those calls.
// The source can be found at cmd/executor-server
type RemoteExecuteAPIExperimentConfig struct {
	Metadata   ExperimentMetadata        `yaml:"metadata"`
	Parameters executor.RemoteExecuteAPI `yaml:"parameters"`
}
type Result struct {
	Name      string      `json:"name"`
	URLResult []URLResult `json:"url_result"`
}

type URLResult struct {
	URL     string `json:"url"`
	Success bool   `json:"success"`
}

func (p *RemoteExecuteAPIExperimentConfig) Type() string {
	return "remote-execute-api"
}

func (p *RemoteExecuteAPIExperimentConfig) Description() string {
	return "Runs a deployment based on a configurable image and then verifies based off of API calls to that image"
}
func (p *RemoteExecuteAPIExperimentConfig) Technique() string {
	return categories.MITRE.Execution.NewContainer.Technique
}
func (p *RemoteExecuteAPIExperimentConfig) Tactic() string {
	return categories.MITRE.Execution.NewContainer.Tactic
}
func (p *RemoteExecuteAPIExperimentConfig) Framework() string {
	return string(categories.Mitre)
}

func (p *RemoteExecuteAPIExperimentConfig) Run(ctx context.Context, experimentConfig *ExperimentConfig) error {
	client, err := k8s.NewClient()
	if err != nil {
		return err
	}
	var config RemoteExecuteAPIExperimentConfig
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err = yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return err
	}

	executorConfig := executor.NewExecutorConfig(
		config.Metadata.Name,
		config.Metadata.Namespace,
		config.Parameters.Image,
		config.Parameters.ImageParameters,
		config.Parameters.ServiceAccountName,
		config.Parameters.Target.Port,
	)

	err = executorConfig.Deploy(ctx, client.Clientset)
	if err != nil {
		return err
	}

	return nil
}

func (p *RemoteExecuteAPIExperimentConfig) Verify(ctx context.Context, experimentConfig *ExperimentConfig) (*verifier.Outcome, error) {
	client, err := k8s.NewClient()
	if err != nil {
		return nil, err
	}
	var config RemoteExecuteAPIExperimentConfig
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err = yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return nil, err
	}

	pf := client.NewPortForwarder(ctx)
	if err != nil {
		return nil, err
	}
	defer pf.Stop()
	forwardedPort, err := pf.Forward(config.Metadata.Namespace, fmt.Sprintf("app=%s", config.Metadata.Name), int(config.Parameters.Target.Port))
	if err != nil {
		return nil, err
	}

	path := config.Parameters.Target.Path
	url := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", pf.Addr(), int32(forwardedPort.Local)),
		Path:   path,
	}

	v := verifier.New(
		config.Metadata.Name,
		config.Description(),
		config.Framework(),
		config.Tactic(),
		config.Technique(),
	)

	result, err := p.retrieveAPIResponse(url.String())
	if err != nil {
		return nil, err
	}

	for _, r := range result.URLResult {
		if r.Success {
			v.Success(r.URL)
		} else if !r.Success {
			v.Fail(r.URL)
		}
	}

	return v.GetOutcome(), nil
}

func (p *RemoteExecuteAPIExperimentConfig) retrieveAPIResponse(url string) (*Result, error) {
	var result Result
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (p *RemoteExecuteAPIExperimentConfig) Cleanup(ctx context.Context, experimentConfig *ExperimentConfig) error {
	client, err := k8s.NewClient()
	if err != nil {
		return err
	}
	clientset := client.Clientset
	var config RemoteExecuteAPIExperimentConfig
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err = yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return err
	}

	executorConfig := executor.NewExecutorConfig(
		config.Metadata.Name,
		config.Metadata.Namespace,
		config.Parameters.Image,
		config.Parameters.ImageParameters,
		config.Parameters.ServiceAccountName,
		config.Parameters.Target.Port,
	)

	err = executorConfig.Cleanup(ctx, clientset)
	if err != nil {
		return err
	}

	return nil
}
