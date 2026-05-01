package prompting

import (
	"strings"
	"testing"

	"github.com/aqcool/langextract/core"
)

func TestPromptBuilderBasic(t *testing.T) {
	template := &PromptTemplateStructured{
		Description: "Extract characters and emotions.",
	}
	builder := NewPromptBuilder(template, core.FormatTypeJSON)

	prompt, err := builder.BuildPrompt("Lady Juliet gazed at the stars.", nil)
	if err != nil {
		t.Fatalf("BuildPrompt failed: %v", err)
	}

	if !strings.Contains(prompt, "Extract characters and emotions.") {
		t.Error("prompt should contain description")
	}
	if !strings.Contains(prompt, "Lady Juliet gazed at the stars.") {
		t.Error("prompt should contain input text")
	}
	if !strings.Contains(prompt, "Q: ") {
		t.Error("prompt should contain Q: prefix")
	}
	if !strings.Contains(prompt, "A: ") {
		t.Error("prompt should contain A: prefix")
	}
}

func TestPromptBuilderWithExamples(t *testing.T) {
	examples := []*core.ExampleData{
		core.NewExampleData(
			"ROMEO. But soft! What light through yonder window breaks?",
			&core.Extraction{
				ExtractionClass: "character",
				ExtractionText:  "ROMEO",
				Attributes:      map[string]interface{}{"emotional_state": "wonder"},
			},
		),
	}

	template := &PromptTemplateStructured{
		Description: "Extract characters.",
		Examples:    examples,
	}
	builder := NewPromptBuilder(template, core.FormatTypeJSON)

	prompt, err := builder.BuildPrompt("Juliet speaks softly.", nil)
	if err != nil {
		t.Fatalf("BuildPrompt failed: %v", err)
	}

	if !strings.Contains(prompt, "Examples") {
		t.Error("prompt should contain examples heading")
	}
	if !strings.Contains(prompt, "ROMEO") {
		t.Error("prompt should contain example data")
	}
	if !strings.Contains(prompt, "character") {
		t.Error("prompt should contain extraction class")
	}
}

func TestPromptBuilderWithAdditionalContext(t *testing.T) {
	template := &PromptTemplateStructured{
		Description: "Extract entities.",
	}
	builder := NewPromptBuilder(template, core.FormatTypeJSON)

	ctx := "This is a medical document."
	prompt, err := builder.BuildPrompt("Patient took Aspirin.", &ctx)
	if err != nil {
		t.Fatalf("BuildPrompt failed: %v", err)
	}

	if !strings.Contains(prompt, "This is a medical document.") {
		t.Error("prompt should contain additional context")
	}
}

func TestFormatHandlerJSONParseOutput(t *testing.T) {
	fh := NewFormatHandler(core.FormatTypeJSON)

	jsonOutput := `{
		"extractions": [
			{"extraction_class": "person", "extraction_text": "Juliet"},
			{"extraction_class": "emotion", "extraction_text": "longing"}
		]
	}`

	extractions, err := fh.ParseOutput(jsonOutput)
	if err != nil {
		t.Fatalf("ParseOutput failed: %v", err)
	}

	if len(extractions) != 2 {
		t.Fatalf("expected 2 extractions, got %d", len(extractions))
	}
	if extractions[0].ExtractionClass != "person" {
		t.Errorf("expected 'person', got '%s'", extractions[0].ExtractionClass)
	}
	if extractions[0].ExtractionText != "Juliet" {
		t.Errorf("expected 'Juliet', got '%s'", extractions[0].ExtractionText)
	}
}

func TestFormatHandlerJSONParseWithAttributes(t *testing.T) {
	fh := NewFormatHandler(core.FormatTypeJSON)

	jsonOutput := `{
		"extractions": [
			{
				"extraction_class": "character",
				"extraction_text": "ROMEO",
				"character_attributes": {"emotional_state": "wonder"}
			}
		]
	}`

	extractions, err := fh.ParseOutput(jsonOutput)
	if err != nil {
		t.Fatalf("ParseOutput failed: %v", err)
	}

	if len(extractions) != 1 {
		t.Fatalf("expected 1 extraction, got %d", len(extractions))
	}
	if extractions[0].Attributes == nil {
		t.Fatal("expected attributes to be set")
	}
	if extractions[0].Attributes["emotional_state"] != "wonder" {
		t.Errorf("expected 'wonder', got '%v'", extractions[0].Attributes["emotional_state"])
	}
}

func TestFormatHandlerYAMLParseOutput(t *testing.T) {
	fh := NewFormatHandler(core.FormatTypeYAML)

	yamlOutput := `extractions:
  - extraction_class: person
    extraction_text: Juliet
  - extraction_class: emotion
    extraction_text: longing
`
	extractions, err := fh.ParseOutput(yamlOutput)
	if err != nil {
		t.Fatalf("ParseOutput failed: %v", err)
	}

	if len(extractions) != 2 {
		t.Fatalf("expected 2 extractions, got %d", len(extractions))
	}
	if extractions[0].ExtractionClass != "person" {
		t.Errorf("expected 'person', got '%s'", extractions[0].ExtractionClass)
	}
}

func TestFormatHandlerEmptyOutput(t *testing.T) {
	fh := NewFormatHandler(core.FormatTypeJSON)

	extractions, err := fh.ParseOutput("")
	if err != nil {
		t.Fatalf("ParseOutput failed: %v", err)
	}
	if extractions != nil {
		t.Error("expected nil extractions for empty output")
	}
}

func TestFormatHandlerInvalidJSON(t *testing.T) {
	fh := NewFormatHandler(core.FormatTypeJSON)

	_, err := fh.ParseOutput("not json at all")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestFormatExtractionExample(t *testing.T) {
	fh := NewFormatHandler(core.FormatTypeJSON)

	extractions := []*core.Extraction{
		{ExtractionClass: "person", ExtractionText: "Juliet",
			Attributes: map[string]interface{}{"mood": "sad"}},
	}

	result, err := fh.FormatExtractionExample(extractions)
	if err != nil {
		t.Fatalf("FormatExtractionExample failed: %v", err)
	}

	if !strings.Contains(result, "person") {
		t.Error("formatted output should contain extraction class")
	}
	if !strings.Contains(result, "Juliet") {
		t.Error("formatted output should contain extraction text")
	}
	if !strings.Contains(result, "person_attributes") {
		t.Error("formatted output should contain attribute key with suffix")
	}
}

func TestFormatExtractionExampleEmpty(t *testing.T) {
	fh := NewFormatHandler(core.FormatTypeJSON)
	result, err := fh.FormatExtractionExample([]*core.Extraction{})
	if err != nil {
		t.Fatalf("FormatExtractionExample failed: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestBuildPrompts(t *testing.T) {
	template := &PromptTemplateStructured{
		Description: "Extract entities.",
	}
	builder := NewPromptBuilder(template, core.FormatTypeJSON)

	texts := []string{"Text one.", "Text two."}
	prompts, err := builder.BuildPrompts(texts, nil)
	if err != nil {
		t.Fatalf("BuildPrompts failed: %v", err)
	}
	if len(prompts) != 2 {
		t.Fatalf("expected 2 prompts, got %d", len(prompts))
	}
	if !strings.Contains(prompts[0], "Text one.") {
		t.Error("first prompt should contain 'Text one.'")
	}
	if !strings.Contains(prompts[1], "Text two.") {
		t.Error("second prompt should contain 'Text two.'")
	}
}
