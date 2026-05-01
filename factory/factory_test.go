package factory

import (
	"testing"

	"github.com/aqcool/langextract/core"
)

func TestNewModelConfig(t *testing.T) {
	config := NewModelConfig("gpt-4o-mini",
		WithProvider("openai"),
		WithProviderKwargs(map[string]interface{}{
			"api_key": "test-key",
		}),
	)

	if config.ModelID != "gpt-4o-mini" {
		t.Errorf("expected 'gpt-4o-mini', got '%s'", config.ModelID)
	}
	if config.Provider != "openai" {
		t.Errorf("expected 'openai', got '%s'", config.Provider)
	}
}

func TestCreateModelNilConfig(t *testing.T) {
	_, err := CreateModel(nil)
	if err == nil {
		t.Error("expected error for nil config")
	}
	if !core.IsInferenceConfigError(err) {
		t.Error("expected InferenceConfigError")
	}
}

func TestCreateModelEmptyModelID(t *testing.T) {
	config := &ModelConfig{}
	_, err := CreateModel(config)
	if err == nil {
		t.Error("expected error for empty model ID")
	}
}

func TestCreateModelUnsupportedProvider(t *testing.T) {
	config := &ModelConfig{
		ModelID:        "test-model",
		Provider:       "unsupported",
		ProviderKwargs: map[string]interface{}{},
	}
	_, err := CreateModel(config)
	if err == nil {
		t.Error("expected error for unsupported provider")
	}
	if !core.IsInferenceConfigError(err) {
		t.Error("expected InferenceConfigError")
	}
}

func TestCreateOpenAIModelNoAPIKey(t *testing.T) {
	config := &ModelConfig{
		ModelID:        "gpt-4o-mini",
		Provider:       "openai",
		ProviderKwargs: map[string]interface{}{},
	}
	_, err := CreateModel(config)
	if err == nil {
		t.Error("expected error when no API key provided")
	}
}

func TestCreateOpenAIModelWithAPIKey(t *testing.T) {
	config := &ModelConfig{
		ModelID:  "gpt-4o-mini",
		Provider: "openai",
		ProviderKwargs: map[string]interface{}{
			"api_key":     "test-key",
			"base_url":    "http://localhost:1234/v1",
			"format_type": core.FormatTypeJSON,
			"max_workers": 5,
		},
	}

	model, err := CreateModel(config)
	if err != nil {
		t.Fatalf("CreateModel failed: %v", err)
	}
	if model == nil {
		t.Fatal("expected non-nil model")
	}
	if model.ModelID() != "gpt-4o-mini" {
		t.Errorf("expected 'gpt-4o-mini', got '%s'", model.ModelID())
	}
	if model.Provider() != "openai" {
		t.Errorf("expected 'openai', got '%s'", model.Provider())
	}
}
