// Package providers provides LLM provider interfaces and implementations.
package providers

import (
	"context"

	"github.com/aqcool/langextract/core"
)

// CallOption is a functional option for LLM call configuration.
type CallOption func(*CallConfig)

// CallConfig holds configuration for LLM calls.
type CallConfig struct {
	Temperature   *float64
	MaxTokens     *int
	TopP          *float64
	StopSequences []string
}

// WithTemperature sets the sampling temperature.
func WithTemperature(temp float64) CallOption {
	return func(c *CallConfig) {
		c.Temperature = &temp
	}
}

// WithMaxTokens sets the maximum number of tokens to generate.
func WithMaxTokens(maxTokens int) CallOption {
	return func(c *CallConfig) {
		c.MaxTokens = &maxTokens
	}
}

// WithTopP sets the top-p sampling parameter.
func WithTopP(topP float64) CallOption {
	return func(c *CallConfig) {
		c.TopP = &topP
	}
}

// WithStopSequences sets the stop sequences.
func WithStopSequences(sequences []string) CallOption {
	return func(c *CallConfig) {
		c.StopSequences = sequences
	}
}

// BaseLanguageModel is an abstract inference interface for managing LLM inference.
// It defines the contract that all LLM providers must implement.
type BaseLanguageModel interface {
	// Call invokes the LLM with the given prompts and returns scored outputs.
	Call(ctx context.Context, prompts []string, opts ...CallOption) ([]core.ScoredOutput, error)
	
	// ModelID returns the model identifier.
	ModelID() string
	
	// Provider returns the provider name.
	Provider() string
	
	// ApplySchema applies a schema instance to this provider.
	// Optional method that providers can implement to store the schema instance
	// for runtime use.
	ApplySchema(schema BaseSchema) error
	
	// Schema returns the current schema instance if one is configured.
	Schema() BaseSchema
	
	// RequiresFenceOutput returns whether this model requires fence output for parsing.
	// Uses explicit override if set, otherwise computes from schema.
	RequiresFenceOutput() bool
	
	// SetFenceOutput sets explicit fence output preference.
	// True to force fences, False to disable, nil for auto.
	SetFenceOutput(fenceOutput *bool)
}

// BaseSchema is the abstract base class for generating structured constraints from examples.
type BaseSchema interface {
	// FromExamples builds a schema instance from example data.
	FromExamples(examples []*core.ExampleData, attributeSuffix string) (BaseSchema, error)
	
	// ToProviderConfig converts schema to provider-specific configuration.
	// Returns a dictionary of provider kwargs.
	ToProviderConfig() map[string]interface{}
	
	// RequiresRawOutput returns whether this schema outputs raw JSON/YAML without fence markers.
	// When True, the provider emits syntactically valid JSON directly.
	// When False, the provider needs fence markers for structure.
	RequiresRawOutput() bool
	
	// ValidateFormat validates format compatibility and warns about issues.
	ValidateFormat(formatType core.FormatType) error
	
	// SyncWithProviderKwargs updates schema state based on provider kwargs.
	// This allows schemas to adjust their behavior based on caller overrides.
	SyncWithProviderKwargs(kwargs map[string]interface{})
}

// FormatModeSchema is a generic schema for providers that support format modes (JSON/YAML).
// This schema doesn't enforce structure, only output format. Useful for providers
// that can guarantee syntactically valid JSON or YAML but don't support field-level constraints.
type FormatModeSchema struct {
	formatType core.FormatType
	format     string
}

// NewFormatModeSchema creates a new FormatModeSchema with the given format type.
func NewFormatModeSchema(formatType core.FormatType) *FormatModeSchema {
	format := "json"
	if formatType == core.FormatTypeYAML {
		format = "yaml"
	}
	return &FormatModeSchema{
		formatType: formatType,
		format:     format,
	}
}

// FromExamples builds a schema instance from example data.
func (s *FormatModeSchema) FromExamples(examples []*core.ExampleData, attributeSuffix string) (BaseSchema, error) {
	// Default to JSON format
	return NewFormatModeSchema(core.FormatTypeJSON), nil
}

// ToProviderConfig converts schema to provider-specific configuration.
func (s *FormatModeSchema) ToProviderConfig() map[string]interface{} {
	return map[string]interface{}{
		"format": s.format,
	}
}

// RequiresRawOutput returns whether this schema outputs raw JSON/YAML without fences.
func (s *FormatModeSchema) RequiresRawOutput() bool {
	return s.format == "json"
}

// ValidateFormat validates format compatibility.
func (s *FormatModeSchema) ValidateFormat(formatType core.FormatType) error {
	// No validation needed for format mode schema
	return nil
}

// SyncWithProviderKwargs updates schema state based on provider kwargs.
func (s *FormatModeSchema) SyncWithProviderKwargs(kwargs map[string]interface{}) {
	if formatType, ok := kwargs["format_type"].(core.FormatType); ok {
		s.formatType = formatType
		if formatType == core.FormatTypeJSON {
			s.format = "json"
		} else {
			s.format = "yaml"
		}
	}
	if format, ok := kwargs["format"].(string); ok {
		s.format = format
		if format == "json" {
			s.formatType = core.FormatTypeJSON
		} else {
			s.formatType = core.FormatTypeYAML
		}
	}
}

// FormatType returns the current format type.
func (s *FormatModeSchema) FormatType() core.FormatType {
	return s.formatType
}

// Format returns the current format string.
func (s *FormatModeSchema) Format() string {
	return s.format
}
