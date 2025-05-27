package experiments

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	dockerClient "github.com/docker/docker/client"
	"github.com/operantai/woodpecker/internal/categories"
	"github.com/operantai/woodpecker/internal/k8s"
	"github.com/operantai/woodpecker/internal/verifier"
	"gopkg.in/yaml.v3"
	"net/http"
	"net/url"
	"time"
)

type LLMDataLeakageExperiment struct {
	Metadata   ExperimentMetadata `yaml:"metadata"`
	Parameters LLMDataLeakage     `yaml:"parameters"`
}

type LLMDataLeakage struct {
	Apis []ExecuteAIAPI `yaml:"apis"`
}

func (p *LLMDataLeakageExperiment) Type() string {
	return "llm-data-leakage"
}

func (p *LLMDataLeakageExperiment) Description() string {
	return "Check whether the LLM AI Model is leaking any sensitive data such as PII data or secrets and keys in its response"
}
func (p *LLMDataLeakageExperiment) Technique() string {
	return categories.MITREATLAS.Exfiltration.LLMDataLeakage.Technique
}
func (p *LLMDataLeakageExperiment) Tactic() string {
	return categories.MITREATLAS.Exfiltration.LLMDataLeakage.Tactic
}
func (p *LLMDataLeakageExperiment) Framework() string {
	return string(categories.MitreAtlas)
}

func (p *LLMDataLeakageExperiment) Run(ctx context.Context, experimentConfig *ExperimentConfig) error {
	var config LLMDataLeakageExperiment
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err := yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return err
	}

	results := make(map[string]ExecuteAIAPIResult)
	var addr string
	var port int
	switch config.Metadata.Namespace {
	case "local":
		client, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
		if err != nil {
			return err
		}
		defer client.Close()
		if !isWoodpeckerAIDockerComponentPresent(ctx, client) {
			return errors.New("Error in checking for woodpecker AI component to run AI experiments. Is it deployed? Deploy with woodpecker component install command.")
		}
		addr = "127.0.0.1"
		port = 8000
	default:
		client, err := k8s.NewClient()
		if err != nil {
			return err
		}
		if !isWoodpeckerAIK8sComponentPresent(ctx, client, config.Metadata.Namespace) {
			return errors.New("Error in checking for woodpecker AI component to run AI experiments. Is it deployed? Deploy with woodpecker component install command.")
		}
		pf := client.NewPortForwarder(ctx)
		if err != nil {
			return err
		}
		defer pf.Stop()
		forwardedPort, err := pf.Forward(config.Metadata.Namespace, fmt.Sprintf("app=%s", WoodpeckerAI), 8000)
		if err != nil {
			return err
		}

		addr = pf.Addr()
		port = int(forwardedPort.Local)
	}

	for _, api := range config.Parameters.Apis {
		url := url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%d", addr, port),
			Path:   "/ai-experiments",
		}

		var requestBody []byte
		if &api.Payload != nil {
			requestBody, err = json.Marshal(api.Payload)
			if err != nil {
				return err
			}
		}
		req, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(requestBody))
		if err != nil {
			return err
		}

		req.Header.Add("Content-type", "application/json")

		response, err := http.DefaultClient.Do(req)
		if err != nil || response.StatusCode != 200 {
			return err
		}
		defer response.Body.Close()
		var apiResponse AIAPIResponse
		err = json.NewDecoder(response.Body).Decode(&apiResponse)
		if err != nil {
			return err
		}

		results[api.Description] = ExecuteAIAPIResult{
			Description:    api.Description,
			ExperimentName: config.Metadata.Name,
			Timestamp:      time.Now(),
			Status:         response.StatusCode,
			Response:       apiResponse,
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
	return nil
}

func (p *LLMDataLeakageExperiment) Verify(ctx context.Context, experimentConfig *ExperimentConfig) (*verifier.Outcome, error) {
	var config LLMDataLeakageExperiment
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

	var aiVerifierOutcome *verifier.AIVerifierOutcome
	for _, rawResult := range rawResults {
		var results map[string]ExecuteAIAPIResult
		err = json.Unmarshal(rawResult, &results)
		if err != nil {
			return nil, fmt.Errorf("Could not parse experiment result: %w", err)
		}

		for _, api := range config.Parameters.Apis {
			result, found := results[api.Description]
			if !found {
				continue
			}
			fail := false
			if api.ExpectedResponse.VerifiedResponseChecks != nil {
				for _, responseCheck := range api.ExpectedResponse.VerifiedResponseChecks {
					if result.Response.VerifiedResponseChecks != nil {
						fail = true
						v.Fail(api.Description)
						if aiVerifierOutcome == nil {
							aiVerifierOutcome = &verifier.AIVerifierOutcome{
								Model:                  result.Response.Model,
								AIApi:                  result.Response.AIApi,
								Prompt:                 result.Response.Prompt,
								APIResponse:            result.Response.APIResponse,
								VerifiedResponseChecks: nil,
							}
						}
						for _, resultCheck := range result.Response.VerifiedResponseChecks {
							if resultCheck.Check == responseCheck.Check {
								aiVerifierOutcome.VerifiedResponseChecks = append(aiVerifierOutcome.VerifiedResponseChecks, verifier.AIVerifierResult{
									Check:      resultCheck.Check,
									Detected:   resultCheck.Detected,
									Score:      resultCheck.Score,
									EntityType: resultCheck.EntityType,
								})
							}
						}

					}
				}
			}
			if !fail {
				v.Success(api.Description)
			}
			v.StoreResultOutputs(api.Description, aiVerifierOutcome)
		}
	}

	return v.GetOutcome(), nil
}

func (p *LLMDataLeakageExperiment) Cleanup(ctx context.Context, experimentConfig *ExperimentConfig) error {
	var config LLMDataLeakageExperiment
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err := yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return err
	}

	if err := removeTempFilesForExperiment(p.Type(), config.Metadata.Name); err != nil {
		return err
	}

	return nil
}
