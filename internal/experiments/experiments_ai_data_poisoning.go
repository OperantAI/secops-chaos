package experiments

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/operantai/woodpecker/internal/categories"
	"github.com/operantai/woodpecker/internal/k8s"
	"github.com/operantai/woodpecker/internal/verifier"
	"gopkg.in/yaml.v3"
	"net/http"
	"net/url"
	"time"
)

type LLMDataPoisoningExperiment struct {
	Metadata   ExperimentMetadata `yaml:"metadata"`
	Parameters LLMDataPoison      `yaml:"parameters"`
}

type LLMDataPoison struct {
	Apis []ExecuteAIAPI `yaml:"apis"`
}

func (p *LLMDataPoisoningExperiment) Type() string {
	return "llm-data-poisoning"
}

func (p *LLMDataPoisoningExperiment) Description() string {
	return "Check whether data or prompts sent to an AI API for training or fine-tuning includes sensitive data"
}
func (p *LLMDataPoisoningExperiment) Technique() string {
	return categories.MITREATLAS.Persistence.PoisonTrainingData.Technique
}
func (p *LLMDataPoisoningExperiment) Tactic() string {
	return categories.MITREATLAS.Persistence.PoisonTrainingData.Tactic
}
func (p *LLMDataPoisoningExperiment) Framework() string {
	return string(categories.MitreAtlas)
}

func (p *LLMDataPoisoningExperiment) Run(ctx context.Context, client *k8s.Client, experimentConfig *ExperimentConfig) error {
	var config LLMDataPoisoningExperiment
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err := yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return err
	}

	if !isWoodpeckerAIComponentPresent(ctx, client, config.Metadata.Namespace) {
		return errors.New("Error in checking for woodpecker AI component to run AI experiments. Is it deployed? Deploy with woodpecker component install command.")
	}
	pf := client.NewPortForwarder(ctx)
	if err != nil {
		return err
	}
	defer pf.Stop()
	forwardedPort, err := pf.Forward(config.Metadata.Namespace, fmt.Sprintf("app=%s", SecopsChaosAi), 8000)
	if err != nil {
		return err
	}
	results := make(map[string]ExecuteAIAPIResult)
	for _, api := range config.Parameters.Apis {

		url := url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%d", pf.Addr(), forwardedPort.Local),
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

func (p *LLMDataPoisoningExperiment) Verify(ctx context.Context, client *k8s.Client, experimentConfig *ExperimentConfig) (*verifier.Outcome, error) {
	var config LLMDataPoisoningExperiment
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
			if api.ExpectedResponse.VerifiedPromptChecks != nil {
				for _, responseCheck := range api.ExpectedResponse.VerifiedPromptChecks {
					if result.Response.VerifiedPromptChecks != nil {
						fail = true
						v.Fail(api.Description)
						if aiVerifierOutcome == nil {
							aiVerifierOutcome = &verifier.AIVerifierOutcome{
								Model:                result.Response.Model,
								AIApi:                result.Response.AIApi,
								Prompt:               result.Response.Prompt,
								APIResponse:          result.Response.APIResponse,
								VerifiedPromptChecks: nil,
							}
						}
						for _, resultCheck := range result.Response.VerifiedPromptChecks {
							if resultCheck.Check == responseCheck.Check {
								aiVerifierOutcome.VerifiedPromptChecks = append(aiVerifierOutcome.VerifiedPromptChecks, verifier.AIVerifierResult{
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

func (p *LLMDataPoisoningExperiment) Cleanup(ctx context.Context, client *k8s.Client, experimentConfig *ExperimentConfig) error {
	var config LLMDataPoisoningExperiment
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
