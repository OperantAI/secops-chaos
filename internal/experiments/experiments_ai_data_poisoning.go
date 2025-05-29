package experiments

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/operantai/woodpecker/internal/categories"
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

func (p *LLMDataPoisoningExperiment) Run(ctx context.Context, experimentConfig *ExperimentConfig) error {
	var config LLMDataPoisoningExperiment
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err := yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return err
	}

	results := make(map[string]ExecuteAIAPIResult)

	verifierAddr, appAddr, err := getAIComponentAddrs(ctx, experimentConfig)
	if err != nil {
		return err
	}

	for _, api := range config.Parameters.Apis {
		appURL := url.URL{
			Scheme: "http",
			Host:   appAddr,
			Path:   "/chat",
		}

		var appRequestBody []byte
		if &api.Payload == nil {
			return errors.New("Payload may not be empty")
		}

		aiAppPayload := AIAppPayload{
			Model:        api.Payload.Model,
			SystemPrompt: api.Payload.SystemPrompt,
			Prompt:       api.Payload.Prompt,
		}

		appRequestBody, err = json.Marshal(aiAppPayload)
		if err != nil {
			return err
		}
		appReq, err := http.NewRequest("POST", appURL.String(), bytes.NewBuffer(appRequestBody))
		if err != nil {
			return err
		}

		appReq.Header.Add("Content-type", "application/json")

		appResponse, err := http.DefaultClient.Do(appReq)
		if err != nil || appResponse.StatusCode != 200 {
			return err
		}
		defer appResponse.Body.Close()

		var ar AIAppResponse
		err = json.NewDecoder(appResponse.Body).Decode(&ar)
		if err != nil {
			return err
		}

		verifierURL := url.URL{
			Scheme: "http",
			Host:   verifierAddr,
			Path:   "/v1/ai-experiments",
		}

		var verifierRequestBody []byte
		verifierRequest := AIAPIPayload{
			Model:                aiAppPayload.Model,
			SystemPrompt:         aiAppPayload.SystemPrompt,
			Prompt:               aiAppPayload.Prompt,
			AIApi:                api.Payload.AIApi,
			Response:             ar.Message,
			VerifyPromptChecks:   api.Payload.VerifyPromptChecks,
			VerifyResponseChecks: api.Payload.VerifyResponseChecks,
		}

		verifierRequestBody, err = json.Marshal(verifierRequest)
		if err != nil {
			return err
		}
		verifierReq, err := http.NewRequest("POST", verifierURL.String(), bytes.NewBuffer(verifierRequestBody))
		if err != nil {
			return err
		}

		verifierReq.Header.Add("Content-type", "application/json")

		verifierResponse, err := http.DefaultClient.Do(verifierReq)
		if err != nil || appResponse.StatusCode != 200 {
			return err
		}
		defer verifierResponse.Body.Close()

		var vr AIVerifierAPIResponse
		err = json.NewDecoder(verifierResponse.Body).Decode(&vr)
		if err != nil {
			return err
		}

		results[api.Description] = ExecuteAIAPIResult{
			Description:    api.Description,
			ExperimentName: config.Metadata.Name,
			Timestamp:      time.Now(),
			Status:         appResponse.StatusCode,
			Response:       vr,
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

func (p *LLMDataPoisoningExperiment) Verify(ctx context.Context, experimentConfig *ExperimentConfig) (*verifier.Outcome, error) {
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

func (p *LLMDataPoisoningExperiment) Cleanup(ctx context.Context, experimentConfig *ExperimentConfig) error {
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
