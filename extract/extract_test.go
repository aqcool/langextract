package extract

import (
	"context"
	"testing"

	"github.com/aqcool/langextract/core"
)

func TestExtractNoPromptDescription(t *testing.T) {
	ctx := context.Background()
	_, err := Extract(ctx, "some text")
	if err == nil {
		t.Error("expected error for missing prompt description")
	}
	if !core.IsInferenceConfigError(err) {
		t.Errorf("expected InferenceConfigError, got %T", err)
	}
}

func TestExtractEmptyText(t *testing.T) {
	ctx := context.Background()
	_, err := Extract(ctx, "", WithPromptDescription("extract entities"))
	if err == nil {
		t.Error("expected error for empty text")
	}
}

func TestExtractNilInput(t *testing.T) {
	ctx := context.Background()
	_, err := Extract(ctx, nil, WithPromptDescription("extract entities"))
	if err == nil {
		t.Error("expected error for nil input")
	}
}

func TestToDocumentsString(t *testing.T) {
	docs, err := toDocuments("hello world")
	if err != nil {
		t.Fatalf("toDocuments failed: %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1 doc, got %d", len(docs))
	}
	if docs[0].Text != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", docs[0].Text)
	}
}

func TestToDocumentsEmptyString(t *testing.T) {
	_, err := toDocuments("")
	if err == nil {
		t.Error("expected error for empty string")
	}
}

func TestToDocumentsDocument(t *testing.T) {
	doc := core.NewDocument("test text")
	docs, err := toDocuments(doc)
	if err != nil {
		t.Fatalf("toDocuments failed: %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1 doc, got %d", len(docs))
	}
}

func TestToDocumentsSlice(t *testing.T) {
	docs := []*core.Document{
		core.NewDocument("text 1"),
		core.NewDocument("text 2"),
	}
	result, err := toDocuments(docs)
	if err != nil {
		t.Fatalf("toDocuments failed: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 docs, got %d", len(result))
	}
}

func TestToDocumentsUnsupportedType(t *testing.T) {
	_, err := toDocuments(123)
	if err == nil {
		t.Error("expected error for unsupported type")
	}
}

func TestDeduplicateExtractions(t *testing.T) {
	start := 5
	extractions := []*core.Extraction{
		{ExtractionClass: "person", ExtractionText: "Juliet", CharInterval: &core.CharInterval{StartPos: &start}},
		{ExtractionClass: "person", ExtractionText: "Juliet", CharInterval: &core.CharInterval{StartPos: &start}}, // duplicate
		{ExtractionClass: "emotion", ExtractionText: "longing"},                                                    // different class
	}

	result := deduplicateExtractions(extractions)
	if len(result) != 2 {
		t.Errorf("expected 2 after dedup, got %d", len(result))
	}
}

func TestDeduplicateExtractionsEmpty(t *testing.T) {
	result := deduplicateExtractions(nil)
	if len(result) != 0 {
		t.Errorf("expected 0, got %d", len(result))
	}
}

func TestValidatePrompt(t *testing.T) {
	err := ValidatePrompt("extract entities", nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidatePromptEmptyDescription(t *testing.T) {
	err := ValidatePrompt("", nil)
	if err == nil {
		t.Error("expected error for empty description")
	}
}

func TestValidatePromptEmptyExampleText(t *testing.T) {
	examples := []*core.ExampleData{
		core.NewExampleData("", &core.Extraction{ExtractionClass: "a"}),
	}
	err := ValidatePrompt("extract", examples)
	if err == nil {
		t.Error("expected error for empty example text")
	}
}

func TestExtractFromFileNotFound(t *testing.T) {
	ctx := context.Background()
	_, err := ExtractFromFile(ctx, "/nonexistent/file.txt", WithPromptDescription("test"))
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestExtractFromURLNotImplemented(t *testing.T) {
	ctx := context.Background()
	_, err := ExtractFromURL(ctx, "http://example.com")
	if err == nil {
		t.Error("expected error for URL extraction")
	}
}

func TestConfigDefaults(t *testing.T) {
	config := &Config{
		ModelID:          DefaultModelID,
		FormatType:       core.FormatTypeJSON,
		MaxCharBuffer:    DefaultMaxCharBuffer,
		MaxWorkers:       DefaultMaxWorkers,
		ExtractionPasses: DefaultExtractionPasses,
	}

	if config.ModelID != "gpt-4o-mini" {
		t.Errorf("expected 'gpt-4o-mini', got '%s'", config.ModelID)
	}
	if config.MaxCharBuffer != 1000 {
		t.Errorf("expected 1000, got %d", config.MaxCharBuffer)
	}
	if config.MaxWorkers != 10 {
		t.Errorf("expected 10, got %d", config.MaxWorkers)
	}
}

func TestOptionFunctions(t *testing.T) {
	config := &Config{}
	temp := 0.5

	WithModelID("test-model")(config)
	WithAPIKey("test-key")(config)
	WithBaseURL("http://localhost")(config)
	WithFormatType(core.FormatTypeYAML)(config)
	WithMaxCharBuffer(500)(config)
	WithTemperature(temp)(config)
	WithMaxWorkers(5)(config)
	WithExtractionPasses(3)(config)
	WithPromptDescription("test prompt")(config)
	WithEnableFuzzyAlign(false)(config)
	WithFuzzyThreshold(0.8)(config)

	if config.ModelID != "test-model" {
		t.Error("WithModelID failed")
	}
	if config.APIKey != "test-key" {
		t.Error("WithAPIKey failed")
	}
	if config.BaseURL != "http://localhost" {
		t.Error("WithBaseURL failed")
	}
	if config.FormatType != core.FormatTypeYAML {
		t.Error("WithFormatType failed")
	}
	if config.MaxCharBuffer != 500 {
		t.Error("WithMaxCharBuffer failed")
	}
	if config.Temperature == nil || *config.Temperature != 0.5 {
		t.Error("WithTemperature failed")
	}
	if config.MaxWorkers != 5 {
		t.Error("WithMaxWorkers failed")
	}
	if config.ExtractionPasses != 3 {
		t.Error("WithExtractionPasses failed")
	}
	if config.PromptDescription != "test prompt" {
		t.Error("WithPromptDescription failed")
	}
	if config.EnableFuzzyAlign != false {
		t.Error("WithEnableFuzzyAlign failed")
	}
	if config.FuzzyThreshold != 0.8 {
		t.Error("WithFuzzyThreshold failed")
	}
}
