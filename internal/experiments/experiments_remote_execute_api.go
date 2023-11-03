package experiments

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/operantai/secops-chaos/internal/categories"
	"github.com/operantai/secops-chaos/internal/executor"
	"github.com/operantai/secops-chaos/internal/k8s"
	"github.com/operantai/secops-chaos/internal/verifier"
	"gopkg.in/yaml.v3"
)

type RemoteExecuteAPIExperimentConfig struct {
	Metadata   ExperimentMetadata `yaml:"metadata"`
	Parameters RemoteExecuteAPI   `yaml:"parameters"`
}

// RemoteExecuteAPI is an experiment that uses the remote executor to check a remote output
// The image must be created independently -- the current default is `alconen/egress_server`, which runs a simple web app on port 4000 that checks http connectivity to
// a few domains ("https://google.com", "https://linkedin.com", "https://openai.com/") and responds with a success based on the success of those calls.
// The source can be found at internal/executor/container_examples/egress_server
type RemoteExecuteAPI struct {
	Image                string `yaml:"image"`
	TargetPort           int32  `yaml:"target_port"`
	Path                 string `yaml:"path"`
	ServiceAccountName   string `yaml:"service_account_name"`
	LocalPort            int32  `yaml:"local_port"`
	ExpectedHTTPResponse int    `yaml:"expected_response"`
	ExpectedBody         string `yaml:"expected_body"`
}

func (p *RemoteExecuteAPIExperimentConfig) Type() string {
	return "remote_execute_api"
}

func (p *RemoteExecuteAPIExperimentConfig) Description() string {
	return "This experiment runs a deployment based on a configurable image and then verifies based off of API calls to that image"
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

func (p *RemoteExecuteAPIExperimentConfig) Run(ctx context.Context, client *k8s.Client, experimentConfig *ExperimentConfig) error {
	var config RemoteExecuteAPIExperimentConfig
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err := yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return err
	}

	executorConfig := executor.NewExecutorConfig(
		config.Metadata.Name,
		config.Metadata.Namespace,
		config.Parameters.Image,
		config.Parameters.ServiceAccountName,
		config.Parameters.TargetPort,
		config.Parameters.LocalPort,
	)

	err = executorConfig.Deploy(ctx, client.Clientset)
	if err != nil {
		return err
	}

	return nil
}

func (p *RemoteExecuteAPIExperimentConfig) Verify(ctx context.Context, client *k8s.Client, experimentConfig *ExperimentConfig) (*verifier.Outcome, error) {
	var config RemoteExecuteAPIExperimentConfig
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err := yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return nil, err
	}

	executorConfig := executor.NewExecutorConfig(
		config.Metadata.Name,
		config.Metadata.Namespace,
		config.Parameters.Image,
		config.Parameters.ServiceAccountName,
		config.Parameters.TargetPort,
		config.Parameters.LocalPort,
	)

	stopChan := make(chan struct{}, 1)
	readyChan := make(chan struct{}, 1)
	errch := make(chan error, 1)
	out, errOut := new(bytes.Buffer), new(bytes.Buffer)

	forwardConfig := executor.PortForwardConfig{
		StopCh:  stopChan,
		ReadyCh: readyChan,
		Out:     out,
		ErrOut:  errOut,
	}

	go func() {
		err := executorConfig.OpenLocalPort(ctx, client, config.Parameters.LocalPort, forwardConfig)
		if err != nil {
			errch <- fmt.Errorf("Local Port Failed to open: %w", err)
		}
	}()

	// Waits until local port is ready or open erros
	select {
	case <-readyChan:
		break
	case err := <-errch:
		return nil, err
	}

	path := config.Parameters.Path
	port := config.Parameters.LocalPort
	url := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", "127.0.0.1", port),
		Path:   path,
	}

	v := verifier.New(
		config.Metadata.Name,
		config.Description(),
		config.Framework(),
		config.Tactic(),
		config.Technique(),
	)

	body, statusCode, err := retrieveAPIResponse(url.String())
	if err != nil {
		return nil, err
	}

	if statusCode == config.Parameters.ExpectedHTTPResponse {
		v.Success("ExpectedStatusCodeRecieved")
	} else {
		v.Fail("ExpectedStatusCodeRecieved")
	}

	if body == config.Parameters.ExpectedBody {
		v.Success("ExpectedBodyRecieved")
	} else {
		v.Fail("ExpectedBodyRecieved")
	}

	close(stopChan)
	return v.GetOutcome(), nil
}

func retrieveAPIResponse(url string) (string, int, error) {
	response, err := http.Get(url)
	if err != nil {
		return "", response.StatusCode, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", response.StatusCode, err
	}

	stringBody := strings.TrimSpace(string(body))
	return stringBody, response.StatusCode, nil
}

func (p *RemoteExecuteAPIExperimentConfig) Cleanup(ctx context.Context, client *k8s.Client, experimentConfig *ExperimentConfig) error {
	clientset := client.Clientset
	var config RemoteExecuteAPIExperimentConfig
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err := yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return err
	}

	executorConfig := executor.NewExecutorConfig(
		config.Metadata.Name,
		config.Metadata.Namespace,
		config.Parameters.Image,
		config.Parameters.ServiceAccountName,
		config.Parameters.TargetPort,
		config.Parameters.LocalPort,
	)

	err = executorConfig.Cleanup(ctx, clientset)
	if err != nil {
		return err
	}

	return nil
}
