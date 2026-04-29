// Package factory provides model factory for creating LLM instances.
package factory

import (
	"fmt"

	"github.com/aqcool/langextract/core"
	"github.com/aqcool/langextract/providers"
	"github.com/aqcool/langextract/providers/openai"
)

// ModelConfig holds configuration for instantiating a language model provider.
type ModelConfig struct {
	ModelID        string
	Provider       string
	ProviderKwargs map[string]interface{}
}

// NewModelConfig creates a new ModelConfig.
func NewModelConfig(modelID string, opts ...ModelConfigOption) *ModelConfig {
	config := &ModelConfig{
		ModelID:        modelID,
		ProviderKwargs: make(map[string]interface{}),
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}

// ModelConfigOption is a functional option for ModelConfig.
type ModelConfigOption func(*ModelConfig)

// WithProvider sets the provider name.
func WithProvider(provider string) ModelConfigOption {
	return func(c *ModelConfig) {
		c.Provider = provider
	}
}

// WithProviderKwargs sets provider-specific keyword arguments.
func WithProviderKwargs(kwargs map[string]interface{}) ModelConfigOption {
	return func(c *ModelConfig) {
		c.ProviderKwargs = kwargs
	}
}

// CreateModel creates a language model instance based on configuration.
func CreateModel(config *ModelConfig) (providers.BaseLanguageModel, error) {
	if config == nil {
		return nil, core.NewInferenceConfigError("model config is nil")
	}

	if config.ModelID == "" {
		return nil, core.NewInferenceConfigError("model_id is required")
	}

	// Determine provider
	provider := config.Provider
	if provider == "" {
		provider = detectProvider(config.ModelID)
	}

	// Create model based on provider
	switch provider {
	case "openai":
		return createOpenAIModel(config)
	default:
		return nil, core.NewInferenceConfigError(
			fmt.Sprintf("unsupported provider: %s", provider),
		)
	}
}

// detectProvider detects the provider from model ID.
func detectProvider(modelID string) string {
	// Default to OpenAI for all models
	return "openai"
}

// createOpenAIModel creates an OpenAI model instance.
func createOpenAIModel(config *ModelConfig) (providers.BaseLanguageModel, error) {
	opts := []openai.Option{}

	// Extract common options from ProviderKwargs
	if apiKey, ok := config.ProviderKwargs["api_key"].(string); ok {
		opts = append(opts, openai.WithAPIKey(apiKey))
	}

	if baseURL, ok := config.ProviderKwargs["base_url"].(string); ok {
		opts = append(opts, openai.WithBaseURL(baseURL))
	}

	if org, ok := config.ProviderKwargs["organization"].(string); ok {
		opts = append(opts, openai.WithOrganization(org))
	}

	if formatType, ok := config.ProviderKwargs["format_type"].(core.FormatType); ok {
		opts = append(opts, openai.WithFormatType(formatType))
	}

	if temp, ok := config.ProviderKwargs["temperature"].(float64); ok {
		opts = append(opts, openai.WithTemperature(temp))
	}

	if maxWorkers, ok := config.ProviderKwargs["max_workers"].(int); ok {
		opts = append(opts, openai.WithMaxWorkers(maxWorkers))
	}

	// Pass remaining kwargs
	if len(config.ProviderKwargs) > 0 {
		opts = append(opts, openai.WithExtraKwargs(config.ProviderKwargs))
	}

	return openai.NewOpenAIModel(config.ModelID, opts...)
}
