// Package extract provides the main extraction API for LangExtract.
package extract

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/aqcool/langextract/chunking"
	"github.com/aqcool/langextract/core"
	"github.com/aqcool/langextract/factory"
	"github.com/aqcool/langextract/prompting"
	"github.com/aqcool/langextract/providers"
	"github.com/aqcool/langextract/resolver"
)

const (
	// DefaultModelID is the default model ID for extraction.
	DefaultModelID = "gpt-4o-mini"
	
	// DefaultMaxCharBuffer is the default maximum character buffer for chunking.
	DefaultMaxCharBuffer = 1000
	
	// DefaultMaxWorkers is the default number of parallel workers.
	DefaultMaxWorkers = 10
	
	// DefaultExtractionPasses is the default number of extraction passes.
	DefaultExtractionPasses = 1
)

// Config holds configuration for extraction.
type Config struct {
	ModelID           string
	APIKey            string
	BaseURL           string
	FormatType        core.FormatType
	MaxCharBuffer     int
	Temperature       *float64
	MaxWorkers        int
	ExtractionPasses  int
	PromptDescription string
	Examples          []*core.ExampleData
	Model             providers.BaseLanguageModel
	EnableFuzzyAlign  bool
	FuzzyThreshold    float64
}

// Option is a functional option for extraction configuration.
type Option func(*Config)

// WithModelID sets the model ID.
func WithModelID(modelID string) Option {
	return func(c *Config) {
		c.ModelID = modelID
	}
}

// WithAPIKey sets the API key.
func WithAPIKey(apiKey string) Option {
	return func(c *Config) {
		c.APIKey = apiKey
	}
}

// WithBaseURL sets the base URL.
func WithBaseURL(baseURL string) Option {
	return func(c *Config) {
		c.BaseURL = baseURL
	}
}

// WithFormatType sets the format type.
func WithFormatType(formatType core.FormatType) Option {
	return func(c *Config) {
		c.FormatType = formatType
	}
}

// WithMaxCharBuffer sets the maximum character buffer.
func WithMaxCharBuffer(maxChars int) Option {
	return func(c *Config) {
		c.MaxCharBuffer = maxChars
	}
}

// WithTemperature sets the sampling temperature.
func WithTemperature(temp float64) Option {
	return func(c *Config) {
		c.Temperature = &temp
	}
}

// WithMaxWorkers sets the maximum number of parallel workers.
func WithMaxWorkers(maxWorkers int) Option {
	return func(c *Config) {
		c.MaxWorkers = maxWorkers
	}
}

// WithExtractionPasses sets the number of extraction passes.
func WithExtractionPasses(passes int) Option {
	return func(c *Config) {
		c.ExtractionPasses = passes
	}
}

// WithPromptDescription sets the prompt description.
func WithPromptDescription(desc string) Option {
	return func(c *Config) {
		c.PromptDescription = desc
	}
}

// WithExamples sets the example data.
func WithExamples(examples []*core.ExampleData) Option {
	return func(c *Config) {
		c.Examples = examples
	}
}

// WithModel sets a pre-configured model instance.
func WithModel(model providers.BaseLanguageModel) Option {
	return func(c *Config) {
		c.Model = model
	}
}

// WithEnableFuzzyAlign sets whether to enable fuzzy alignment.
func WithEnableFuzzyAlign(enable bool) Option {
	return func(c *Config) {
		c.EnableFuzzyAlign = enable
	}
}

// WithFuzzyThreshold sets the fuzzy alignment threshold.
func WithFuzzyThreshold(threshold float64) Option {
	return func(c *Config) {
		c.FuzzyThreshold = threshold
	}
}

// Extract extracts structured information from text using LLM.
// This is the main entry point for LangExtract.
func Extract(ctx context.Context, textOrDocuments interface{}, opts ...Option) ([]*core.AnnotatedDocument, error) {
	// Initialize config with defaults
	config := &Config{
		ModelID:          DefaultModelID,
		FormatType:       core.FormatTypeJSON,
		MaxCharBuffer:    DefaultMaxCharBuffer,
		MaxWorkers:       DefaultMaxWorkers,
		ExtractionPasses: DefaultExtractionPasses,
		EnableFuzzyAlign: true,
		FuzzyThreshold:   resolver.FuzzyAlignmentMinThreshold,
	}

	// Apply options
	for _, opt := range opts {
		opt(config)
	}

	// Validate configuration
	if config.PromptDescription == "" {
		return nil, core.NewInferenceConfigError("prompt_description is required")
	}

	// Create or use provided model
	model := config.Model
	if model == nil {
		var err error
		model, err = createModel(config)
		if err != nil {
			return nil, err
		}
	}

	// Convert input to documents
	documents, err := toDocuments(textOrDocuments)
	if err != nil {
		return nil, err
	}

	// Create prompt builder
	template := &prompting.PromptTemplateStructured{
		Description: config.PromptDescription,
		Examples:    config.Examples,
	}
	promptBuilder := prompting.NewPromptBuilder(template, config.FormatType)

	// Create chunker
	chunker := chunking.NewTextChunker(
		chunking.WithMaxCharBuffer(config.MaxCharBuffer),
	)

	// Create resolver
	resolverInstance := resolver.NewResolver(
		resolver.WithFormatType(config.FormatType),
	)

	// Process documents
	results := make([]*core.AnnotatedDocument, len(documents))
	for i, doc := range documents {
		annotatedDoc, err := processDocument(
			ctx,
			doc,
			model,
			promptBuilder,
			chunker,
			resolverInstance,
			config,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to process document %s: %w", doc.DocumentID, err)
		}
		results[i] = annotatedDoc
	}

	return results, nil
}

// createModel creates a model instance from configuration.
func createModel(config *Config) (providers.BaseLanguageModel, error) {
	// Resolve API key
	apiKey := config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			apiKey = os.Getenv("LANGEXTRACT_API_KEY")
		}
	}

	// Build provider kwargs
	kwargs := map[string]interface{}{
		"api_key":     apiKey,
		"format_type": config.FormatType,
		"max_workers": config.MaxWorkers,
	}

	if config.BaseURL != "" {
		kwargs["base_url"] = config.BaseURL
	}

	if config.Temperature != nil {
		kwargs["temperature"] = *config.Temperature
	}

	// Create model config
	modelConfig := factory.NewModelConfig(
		config.ModelID,
		factory.WithProvider("openai"),
		factory.WithProviderKwargs(kwargs),
	)

	return factory.CreateModel(modelConfig)
}

// toDocuments converts input to documents.
func toDocuments(input interface{}) ([]*core.Document, error) {
	switch v := input.(type) {
	case string:
		if v == "" {
			return nil, core.NewInvalidDocumentError("text is empty")
		}
		return []*core.Document{core.NewDocument(v)}, nil

	case *core.Document:
		return []*core.Document{v}, nil

	case []*core.Document:
		return v, nil

	default:
		return nil, core.NewInvalidDocumentError(
			fmt.Sprintf("unsupported input type: %T", input),
		)
	}
}

// processDocument processes a single document.
func processDocument(
	ctx context.Context,
	document *core.Document,
	model providers.BaseLanguageModel,
	promptBuilder *prompting.PromptBuilder,
	chunker *chunking.TextChunker,
	resolverInstance *resolver.Resolver,
	config *Config,
) (*core.AnnotatedDocument, error) {
	slog.Debug("Processing document", "id", document.DocumentID)

	// Chunk document
	chunks, err := chunker.ChunkDocument(document)
	if err != nil {
		return nil, fmt.Errorf("failed to chunk document: %w", err)
	}

	slog.Debug("Document chunked", "chunks", len(chunks))

	// Process each chunk
	var allExtractions []*core.Extraction

	for pass := 0; pass < config.ExtractionPasses; pass++ {
		slog.Debug("Extraction pass", "pass", pass+1, "total", config.ExtractionPasses)

		for chunkIdx, chunk := range chunks {
			chunkText, err := chunk.ChunkText()
			if err != nil {
				return nil, fmt.Errorf("failed to get chunk text: %w", err)
			}

			// Build prompt
			prompt, err := promptBuilder.BuildPrompt(chunkText, chunk.AdditionalContext())
			if err != nil {
				return nil, fmt.Errorf("failed to build prompt: %w", err)
			}

			// Call LLM
			outputs, err := model.Call(ctx, []string{prompt})
			if err != nil {
				return nil, fmt.Errorf("LLM call failed: %w", err)
			}

			if len(outputs) == 0 || outputs[0].Output == nil {
				slog.Warn("No output from LLM", "chunk", chunkIdx)
				continue
			}

			// Parse output
			extractions, err := resolverInstance.Resolve(*outputs[0].Output)
			if err != nil {
				slog.Warn("Failed to parse output", "error", err, "chunk", chunkIdx)
				continue
			}

			// Align extractions with source text
			charOffset := 0
			if chunk.CharInterval != nil {
				if chunk.CharInterval.StartPos != nil {
					charOffset = *chunk.CharInterval.StartPos
				}
			}

			tokenOffset := chunk.TokenInterval.StartIndex

			alignOpts := []resolver.AlignOption{
				resolver.WithEnableFuzzyAlignment(config.EnableFuzzyAlign),
				resolver.WithFuzzyAlignmentThreshold(config.FuzzyThreshold),
			}

			alignedExtractions := resolverInstance.Align(
				extractions,
				document.Text,
				tokenOffset,
				&charOffset,
				alignOpts...,
			)

			allExtractions = append(allExtractions, alignedExtractions...)
		}
	}

	// Deduplicate extractions
	allExtractions = deduplicateExtractions(allExtractions)

	// Create annotated document
	annotatedDoc := core.NewAnnotatedDocument(
		core.WithAnnotatedDocumentID(document.DocumentID),
		core.WithAnnotatedDocumentText(document.Text),
		core.WithExtractions(allExtractions),
	)

	return annotatedDoc, nil
}

// deduplicateExtractions removes duplicate extractions based on text and position.
func deduplicateExtractions(extractions []*core.Extraction) []*core.Extraction {
	if len(extractions) == 0 {
		return extractions
	}

	seen := make(map[string]bool)
	result := make([]*core.Extraction, 0, len(extractions))

	for _, ext := range extractions {
		// Create a key based on class and text
		key := fmt.Sprintf("%s:%s", ext.ExtractionClass, ext.ExtractionText)
		
		// Add position to key if available
		if ext.CharInterval != nil && ext.CharInterval.StartPos != nil {
			key += fmt.Sprintf(":%d", *ext.CharInterval.StartPos)
		}

		if !seen[key] {
			seen[key] = true
			result = append(result, ext)
		}
	}

	return result
}

// ExtractFromURL extracts information from a URL.
func ExtractFromURL(ctx context.Context, url string, opts ...Option) ([]*core.AnnotatedDocument, error) {
	// This is a placeholder - in real implementation, you would fetch the URL content
	return nil, fmt.Errorf("URL extraction not implemented")
}

// ExtractFromFile extracts information from a file.
func ExtractFromFile(ctx context.Context, path string, opts ...Option) ([]*core.AnnotatedDocument, error) {
	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return Extract(ctx, string(content), opts...)
}

// ValidatePrompt validates the prompt description and examples.
func ValidatePrompt(description string, examples []*core.ExampleData) error {
	if description == "" {
		return core.NewInferenceConfigError("prompt description is required")
	}

	// Check that examples have valid text
	for i, example := range examples {
		if strings.TrimSpace(example.Text) == "" {
			return core.NewInferenceConfigError(
				fmt.Sprintf("example %d has empty text", i),
			)
		}
	}

	return nil
}
