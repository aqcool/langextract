// Package prompting provides prompt building functionality for LLM inference.
package prompting

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aqcool/langextract/core"
	"gopkg.in/yaml.v3"
)

// PromptTemplateStructured represents a structured prompt template for few-shot examples.
type PromptTemplateStructured struct {
	Description string              `json:"description" yaml:"description"`
	Examples    []*core.ExampleData `json:"examples" yaml:"examples"`
}

// QAPromptGenerator generates question-answer prompts from the provided template.
type QAPromptGenerator struct {
	Template       *PromptTemplateStructured
	FormatHandler  *FormatHandler
	ExamplesHeading string
	QuestionPrefix  string
	AnswerPrefix    string
}

// NewQAPromptGenerator creates a new QAPromptGenerator.
func NewQAPromptGenerator(template *PromptTemplateStructured, formatType core.FormatType) *QAPromptGenerator {
	return &QAPromptGenerator{
		Template:        template,
		FormatHandler:   NewFormatHandler(formatType),
		ExamplesHeading: "Examples",
		QuestionPrefix:  "Q: ",
		AnswerPrefix:    "A: ",
	}
}

// FormatExampleAsText formats a single example for the prompt.
func (gen *QAPromptGenerator) FormatExampleAsText(example *core.ExampleData) (string, error) {
	question := example.Text
	answer, err := gen.FormatHandler.FormatExtractionExample(example.Extractions)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s\n%s%s\n", gen.QuestionPrefix, question, gen.AnswerPrefix, answer), nil
}

// Render generates a text representation of the prompt.
func (gen *QAPromptGenerator) Render(question string, additionalContext *string) (string, error) {
	var promptLines []string

	// Add description
	promptLines = append(promptLines, gen.Template.Description)

	// Add additional context if provided
	if additionalContext != nil && *additionalContext != "" {
		promptLines = append(promptLines, *additionalContext)
	}

	// Add examples
	if len(gen.Template.Examples) > 0 {
		promptLines = append(promptLines, gen.ExamplesHeading)
		for _, ex := range gen.Template.Examples {
			exampleText, err := gen.FormatExampleAsText(ex)
			if err != nil {
				return "", err
			}
			promptLines = append(promptLines, exampleText)
		}
	}

	// Add question
	promptLines = append(promptLines, fmt.Sprintf("%s%s", gen.QuestionPrefix, question))
	promptLines = append(promptLines, gen.AnswerPrefix)

	return strings.Join(promptLines, "\n"), nil
}

// PromptBuilder builds prompts for text chunks using a QAPromptGenerator.
type PromptBuilder struct {
	Generator *QAPromptGenerator
}

// NewPromptBuilder creates a new PromptBuilder.
func NewPromptBuilder(template *PromptTemplateStructured, formatType core.FormatType) *PromptBuilder {
	return &PromptBuilder{
		Generator: NewQAPromptGenerator(template, formatType),
	}
}

// BuildPrompt builds a prompt for the given chunk.
func (pb *PromptBuilder) BuildPrompt(chunkText string, additionalContext *string) (string, error) {
	return pb.Generator.Render(chunkText, additionalContext)
}

// BuildPrompts builds prompts for multiple chunks.
func (pb *PromptBuilder) BuildPrompts(chunkTexts []string, additionalContext *string) ([]string, error) {
	prompts := make([]string, len(chunkTexts))
	for i, chunkText := range chunkTexts {
		prompt, err := pb.BuildPrompt(chunkText, additionalContext)
		if err != nil {
			return nil, fmt.Errorf("failed to build prompt for chunk %d: %w", i, err)
		}
		prompts[i] = prompt
	}
	return prompts, nil
}

// FormatHandler handles formatting of extractions for prompts.
type FormatHandler struct {
	formatType core.FormatType
}

// NewFormatHandler creates a new FormatHandler.
func NewFormatHandler(formatType core.FormatType) *FormatHandler {
	return &FormatHandler{
		formatType: formatType,
	}
}

// FormatExtractionExample formats extractions as a string for the prompt.
func (fh *FormatHandler) FormatExtractionExample(extractions []*core.Extraction) (string, error) {
	if len(extractions) == 0 {
		return "", nil
	}

	// Build extraction data structure
	data := map[string]interface{}{
		core.ExtractionsKey: fh.extractionsToMap(extractions),
	}

	var result []byte
	var err error

	if fh.formatType == core.FormatTypeJSON {
		result, err = json.MarshalIndent(data, "", "  ")
	} else {
		result, err = yaml.Marshal(data)
	}

	if err != nil {
		return "", fmt.Errorf("failed to format extractions: %w", err)
	}

	return string(result), nil
}

// extractionsToMap converts extractions to a map structure.
func (fh *FormatHandler) extractionsToMap(extractions []*core.Extraction) []map[string]interface{} {
	result := make([]map[string]interface{}, len(extractions))
	for i, ext := range extractions {
		extMap := map[string]interface{}{
			"extraction_class": ext.ExtractionClass,
			"extraction_text":  ext.ExtractionText,
		}

		if ext.Attributes != nil && len(ext.Attributes) > 0 {
			// Add attributes with suffix
			attrKey := ext.ExtractionClass + core.AttributeSuffix
			extMap[attrKey] = ext.Attributes
		}

		result[i] = extMap
	}
	return result
}

// ParseOutput parses LLM output into extractions.
func (fh *FormatHandler) ParseOutput(output string) ([]*core.Extraction, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil, nil
	}

	var data map[string]interface{}
	var err error

	if fh.formatType == core.FormatTypeJSON {
		err = json.Unmarshal([]byte(output), &data)
	} else {
		err = yaml.Unmarshal([]byte(output), &data)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse output: %w", err)
	}

	return fh.parseExtractionsFromMap(data)
}

// parseExtractionsFromMap extracts extractions from a map structure.
func (fh *FormatHandler) parseExtractionsFromMap(data map[string]interface{}) ([]*core.Extraction, error) {
	extractionsRaw, ok := data[core.ExtractionsKey]
	if !ok {
		return nil, nil
	}

	extractionsList, ok := extractionsRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("extractions is not a list")
	}

	extractions := make([]*core.Extraction, 0, len(extractionsList))
	for _, extRaw := range extractionsList {
		extMap, ok := extRaw.(map[string]interface{})
		if !ok {
			continue
		}

		extraction := &core.Extraction{}

		if class, ok := extMap["extraction_class"].(string); ok {
			extraction.ExtractionClass = class
		}

		if text, ok := extMap["extraction_text"].(string); ok {
			extraction.ExtractionText = text
		}

		// Extract attributes with suffix
		if extraction.ExtractionClass != "" {
			attrKey := extraction.ExtractionClass + core.AttributeSuffix
			if attrs, ok := extMap[attrKey]; ok {
				if attrMap, ok := attrs.(map[string]interface{}); ok {
					extraction.Attributes = attrMap
				}
			}
		}

		extractions = append(extractions, extraction)
	}

	return extractions, nil
}

// ReadPromptTemplateFromFile reads a structured prompt template from a file.
func ReadPromptTemplateFromFile(path string, formatType core.FormatType) (*PromptTemplateStructured, error) {
	// This is a placeholder - in real implementation, you would read from file
	// For now, return an error indicating file reading is not implemented
	return nil, fmt.Errorf("file reading not implemented")
}

// ParsePromptTemplate parses a prompt template from string content.
func ParsePromptTemplate(content string, formatType core.FormatType) (*PromptTemplateStructured, error) {
	template := &PromptTemplateStructured{}
	var err error

	if formatType == core.FormatTypeJSON {
		err = json.Unmarshal([]byte(content), template)
	} else {
		err = yaml.Unmarshal([]byte(content), template)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse prompt template: %w", err)
	}

	return template, nil
}
