package experiments

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/operantai/secops-chaos/internal/categories"
	"github.com/operantai/secops-chaos/internal/k8s"
	"github.com/operantai/secops-chaos/internal/verifier"
	"gopkg.in/yaml.v3"
)

type ExecuteAPIExperimentConfig struct {
	Metadata   ExperimentMetadata `yaml:"metadata"`
	Parameters ExecuteAPI         `yaml:"parameters"`
}

type ExecuteAPI struct {
	Targets []ExecuteAPITargets `yaml:"targets"`
}

type ExecuteAPITargets struct {
	Target   string              `yaml:"target"`
	Port     int                 `yaml:"port"`
	Payloads []ExecuteAPIPayload `yaml:"payload"`
}

type ExecuteAPIPayload struct {
	Description      string            `yaml:"description"`
	Path             string            `yaml:"path"`
	Method           string            `yaml:"method"`
	Headers          map[string]string `yaml:"headers"`
	Payload          string            `yaml:"payload"`
	ExpectedResponse string            `yaml:"expected_response"`
}

type ExecuteAPIResult struct {
	ExperimentName string    `json:"experiment_name"`
	Description    string    `json:"description"`
	Timestamp      time.Time `json:"timestamp"`
	Status         int       `json:"status"`
	Response       string    `json:"byte"`
}

func (p *ExecuteAPIExperimentConfig) Type() string {
	return "execute_api"
}

func (p *ExecuteAPIExperimentConfig) Description() string {
	return "This experiment port forwards to a service running in Kubernetes and issues API calls to that service"
}
func (p *ExecuteAPIExperimentConfig) Technique() string {
	return categories.MITRE.Execution.ApplicationExploit.Technique
}
func (p *ExecuteAPIExperimentConfig) Tactic() string {
	return categories.MITRE.Execution.ApplicationExploit.Tactic
}
func (p *ExecuteAPIExperimentConfig) Framework() string {
	return string(categories.Mitre)
}

func (p *ExecuteAPIExperimentConfig) Run(ctx context.Context, client *k8s.Client, experimentConfig *ExperimentConfig) error {
	var config ExecuteAPIExperimentConfig
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err := yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return err
	}

	for _, target := range config.Parameters.Targets {
		pf := client.NewPortForwarder(ctx)
		if err != nil {
			return err
		}
		defer pf.Stop()
		forwardedPort, err := pf.Forward(config.Metadata.Namespace, fmt.Sprintf("app=%s", target.Target), target.Port)
		if err != nil {
			return err
		}
		results := make(map[string]ExecuteAPIResult)
		for _, payload := range target.Payloads {
			url := url.URL{
				Scheme: "http",
				Host:   fmt.Sprintf("%s:%d", pf.Addr(), forwardedPort),
				Path:   payload.Path,
			}

			req, err := http.NewRequest(payload.Method, url.String(), nil)
			if err != nil {
				return err
			}

			for k, v := range payload.Headers {
				req.Header.Add(k, v)
			}

			response, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				return err
			}

			results[payload.Description] = ExecuteAPIResult{
				Description:    payload.Description,
				ExperimentName: config.Metadata.Name,
				Timestamp:      time.Now(),
				Status:         response.StatusCode,
				Response:       string(body),
			}
		}

		resultJSON, err := json.Marshal(results)
		if err != nil {
			return fmt.Errorf("Failed to marshal experiment results: %w", err)
		}

		file, err := createTempFile(p.Type(), config.Metadata.Name)
		if err != nil {
			return fmt.Errorf("Unable to create file cache for experiment results %w", err)
		}

		_, err = file.Write(resultJSON)
		if err != nil {
			return fmt.Errorf("Failed to write experiment results: %w", err)
		}
	}

	return nil
}

func (p *ExecuteAPIExperimentConfig) Verify(ctx context.Context, client *k8s.Client, experimentConfig *ExperimentConfig) (*verifier.Outcome, error) {
	var config ExecuteAPIExperimentConfig
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

	rawResults, err := getTempFileContentsForExperiment(p.Type(), config.Metadata.Name)
	if err != nil {
		return nil, fmt.Errorf("Could not fetch experiment results: %w", err)
	}

	for _, rawResult := range rawResults {
		var results map[string]ExecuteAPIResult
		err = json.Unmarshal(rawResult, &results)
		if err != nil {
			return nil, fmt.Errorf("Could not parse experiment result: %w", err)
		}

		for _, target := range config.Parameters.Targets {
			for _, payload := range target.Payloads {
				result, found := results[payload.Description]
				if !found {
					continue
				}
				if result.Response != payload.ExpectedResponse {
					v.Fail(payload.Description)
					continue
				}
				v.Success(payload.Description)
			}
		}
	}

	return v.GetOutcome(), nil
}

func (p *ExecuteAPIExperimentConfig) Cleanup(ctx context.Context, client *k8s.Client, experimentConfig *ExperimentConfig) error {
	var config RemoteExecuteAPIExperimentConfig
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err := yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return err
	}

	if err := os.Remove("test.txt"); err != nil {
		return err
	}

	return nil
}
