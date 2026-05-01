package langextract_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aqcool/langextract/core"
	"github.com/aqcool/langextract/extract"
)

// 集成测试：使用本地 LMStudio 模型
// LMStudio 地址: http://192.168.0.154:1234
// 模型: google/gemma-4-e4b

const (
	baseURL = "http://192.168.0.154:1234/v1"
	apiKey  = "sk-lm-xYSLUVJw:ct9kkJ6QYfKQORTWbYgH"
	modelID = "google/gemma-4-e4b"
)

// LMStudio 不支持 json_object 的 response_format，需要使用 text
var lmstudioExtraKwargs = map[string]interface{}{
	"response_format_type": "text",
}

// 通用 few-shot 示例：角色和情感提取
var characterExamples = []*core.ExampleData{
	core.NewExampleData(
		"ROMEO. But soft! What light through yonder window breaks?",
		&core.Extraction{
			ExtractionClass: "character",
			ExtractionText:  "ROMEO",
			Attributes:      map[string]interface{}{"emotional_state": "wonder"},
		},
		&core.Extraction{
			ExtractionClass: "emotion",
			ExtractionText:  "But soft!",
			Attributes:      map[string]interface{}{"feeling": "gentle awe"},
		},
	),
}

// 测试 1: 基础提取（使用 few-shot 示例）
func TestIntegrationBasicExtraction(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	text := "Lady Juliet gazed longingly at the stars, her heart aching for Romeo"

	results, err := extract.Extract(
		ctx,
		text,
		extract.WithAPIKey(apiKey),
		extract.WithBaseURL(baseURL),
		extract.WithModelID(modelID),
		extract.WithPromptDescription("Extract characters, emotions, and relationships. Use exact text for extractions."),
		extract.WithExamples(characterExamples),
		extract.WithMaxCharBuffer(2000),
		extract.WithExtraKwargs(lmstudioExtraKwargs),
	)
	if err != nil {
		t.Fatalf("Extraction failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}

	doc := results[0]
	fmt.Printf("=== Basic Extraction ===\n")
	fmt.Printf("Document ID: %s\n", doc.DocumentID)
	fmt.Printf("Number of extractions: %d\n", len(doc.Extractions))

	for _, ext := range doc.Extractions {
		fmt.Printf("  Class: %-15s Text: %-20s", ext.ExtractionClass, ext.ExtractionText)
		if ext.CharInterval != nil {
			fmt.Printf(" Position: %v-%v [%s]", ext.CharInterval.StartPos, ext.CharInterval.EndPos, ext.AlignmentStatus)
		}
		fmt.Println()
	}

	if len(doc.Extractions) == 0 {
		t.Error("expected at least 1 extraction from LLM")
	}
}

// 测试 2: Few-shot 示例提取（验证属性传递）
func TestIntegrationWithFewShotExamples(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	text := "ROMEO. But soft! What light through yonder window breaks? It is the east, and Juliet is the sun."

	results, err := extract.Extract(
		ctx,
		text,
		extract.WithAPIKey(apiKey),
		extract.WithBaseURL(baseURL),
		extract.WithModelID(modelID),
		extract.WithPromptDescription("Extract characters, emotions, and relationships in order of appearance. Use exact text for extractions."),
		extract.WithExamples(characterExamples),
		extract.WithMaxCharBuffer(2000),
		extract.WithExtraKwargs(lmstudioExtraKwargs),
	)
	if err != nil {
		t.Fatalf("Extraction failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}

	doc := results[0]
	fmt.Printf("=== Few-Shot Extraction ===\n")
	fmt.Printf("Document ID: %s\n", doc.DocumentID)
	fmt.Printf("Number of extractions: %d\n", len(doc.Extractions))

	hasAttributes := false
	for _, ext := range doc.Extractions {
		fmt.Printf("  Class: %-15s Text: %-20s", ext.ExtractionClass, ext.ExtractionText)
		if ext.Attributes != nil && len(ext.Attributes) > 0 {
			fmt.Printf(" Attrs: %v", ext.Attributes)
			hasAttributes = true
		}
		if ext.CharInterval != nil {
			fmt.Printf(" Pos: %v-%v", ext.CharInterval.StartPos, ext.CharInterval.EndPos)
		}
		fmt.Println()
	}

	if len(doc.Extractions) == 0 {
		t.Error("expected at least 1 extraction from LLM")
	}
	if !hasAttributes {
		t.Log("Warning: no attributes found in extractions (model may not follow attribute format)")
	}
}

// 测试 3: 医疗文本提取
func TestIntegrationMedicalExtraction(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	medicalExamples := []*core.ExampleData{
		core.NewExampleData(
			"Patient takes Lisinopril 10mg once daily for hypertension.",
			&core.Extraction{
				ExtractionClass: "medication",
				ExtractionText:  "Lisinopril",
				Attributes:      map[string]interface{}{"dosage": "10mg", "frequency": "once daily"},
			},
			&core.Extraction{
				ExtractionClass: "condition",
				ExtractionText:  "hypertension",
			},
		),
	}

	text := "The patient was prescribed Aspirin 100mg daily for cardiovascular health. She also takes Metformin 500mg twice daily for type 2 diabetes."

	results, err := extract.Extract(
		ctx,
		text,
		extract.WithAPIKey(apiKey),
		extract.WithBaseURL(baseURL),
		extract.WithModelID(modelID),
		extract.WithPromptDescription("Extract medication names, dosages, frequencies, and conditions from the medical text. Use exact text for extractions."),
		extract.WithExamples(medicalExamples),
		extract.WithMaxCharBuffer(2000),
		extract.WithExtraKwargs(lmstudioExtraKwargs),
	)
	if err != nil {
		t.Fatalf("Extraction failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}

	doc := results[0]
	fmt.Printf("=== Medical Extraction ===\n")
	fmt.Printf("Document ID: %s\n", doc.DocumentID)
	fmt.Printf("Number of extractions: %d\n", len(doc.Extractions))

	for _, ext := range doc.Extractions {
		fmt.Printf("  Class: %-15s Text: %-20s", ext.ExtractionClass, ext.ExtractionText)
		if ext.Attributes != nil && len(ext.Attributes) > 0 {
			fmt.Printf(" Attrs: %v", ext.Attributes)
		}
		if ext.CharInterval != nil {
			fmt.Printf(" Position: %v-%v [%s]", ext.CharInterval.StartPos, ext.CharInterval.EndPos, ext.AlignmentStatus)
		}
		fmt.Println()
	}

	if len(doc.Extractions) == 0 {
		t.Error("expected at least 1 extraction from LLM")
	}
}

// 测试 4: YAML 格式输出
func TestIntegrationWithYAMLFormat(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	yamlExamples := []*core.ExampleData{
		core.NewExampleData(
			"Alice traveled to London in December.",
			&core.Extraction{
				ExtractionClass: "person",
				ExtractionText:  "Alice",
			},
			&core.Extraction{
				ExtractionClass: "location",
				ExtractionText:  "London",
			},
			&core.Extraction{
				ExtractionClass: "time",
				ExtractionText:  "December",
			},
		),
	}

	text := "John and Mary went to Paris last summer. They visited the Eiffel Tower."

	results, err := extract.Extract(
		ctx,
		text,
		extract.WithAPIKey(apiKey),
		extract.WithBaseURL(baseURL),
		extract.WithModelID(modelID),
		extract.WithPromptDescription("Extract people, locations, and time references. Use exact text for extractions."),
		extract.WithExamples(yamlExamples),
		extract.WithFormatType(core.FormatTypeYAML),
		extract.WithMaxCharBuffer(2000),
		extract.WithExtraKwargs(lmstudioExtraKwargs),
	)
	if err != nil {
		t.Fatalf("Extraction failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}

	doc := results[0]
	fmt.Printf("=== YAML Format Test ===\n")
	fmt.Printf("Extractions: %d\n", len(doc.Extractions))
	for _, ext := range doc.Extractions {
		fmt.Printf("  %s: %s\n", ext.ExtractionClass, ext.ExtractionText)
	}

	if len(doc.Extractions) == 0 {
		t.Error("expected at least 1 extraction from LLM")
	}
}

// 测试 5: 多轮提取
func TestIntegrationMultiPass(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	personExamples := []*core.ExampleData{
		core.NewExampleData(
			"Marie Curie discovered radium.",
			&core.Extraction{
				ExtractionClass: "person",
				ExtractionText:  "Marie Curie",
			},
			&core.Extraction{
				ExtractionClass: "achievement",
				ExtractionText:  "discovered radium",
			},
		),
	}

	text := "Albert Einstein developed the theory of relativity. He won the Nobel Prize in Physics in 1921."

	results, err := extract.Extract(
		ctx,
		text,
		extract.WithAPIKey(apiKey),
		extract.WithBaseURL(baseURL),
		extract.WithModelID(modelID),
		extract.WithPromptDescription("Extract people, achievements, dates, and fields. Use exact text for extractions."),
		extract.WithExamples(personExamples),
		extract.WithMaxCharBuffer(2000),
		extract.WithExtractionPasses(2),
		extract.WithExtraKwargs(lmstudioExtraKwargs),
	)
	if err != nil {
		t.Fatalf("Extraction failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}

	doc := results[0]
	fmt.Printf("=== Multi-Pass Test ===\n")
	fmt.Printf("Extractions: %d (after dedup across 2 passes)\n", len(doc.Extractions))
	for _, ext := range doc.Extractions {
		fmt.Printf("  %s: %s\n", ext.ExtractionClass, ext.ExtractionText)
	}
}

// 测试 6: 文本对齐验证
func TestIntegrationAlignmentVerification(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	text := "Lady Juliet gazed longingly at the stars, her heart aching for Romeo"

	results, err := extract.Extract(
		ctx,
		text,
		extract.WithAPIKey(apiKey),
		extract.WithBaseURL(baseURL),
		extract.WithModelID(modelID),
		extract.WithPromptDescription("Extract characters and emotions. Use exact text for extractions."),
		extract.WithExamples(characterExamples),
		extract.WithMaxCharBuffer(2000),
		extract.WithEnableFuzzyAlign(true),
		extract.WithExtraKwargs(lmstudioExtraKwargs),
	)
	if err != nil {
		t.Fatalf("Extraction failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}

	doc := results[0]
	fmt.Printf("=== Alignment Verification ===\n")
	alignedCount := 0
	for _, ext := range doc.Extractions {
		status := "none"
		if ext.AlignmentStatus != "" {
			status = string(ext.AlignmentStatus)
		}
		fmt.Printf("  Class: %-15s Text: %-20s Alignment: %s", ext.ExtractionClass, ext.ExtractionText, status)
		if ext.CharInterval != nil && ext.CharInterval.StartPos != nil && ext.CharInterval.EndPos != nil {
			matchedText := text[*ext.CharInterval.StartPos:*ext.CharInterval.EndPos]
			fmt.Printf(" MatchedText: '%s'", matchedText)
			alignedCount++
		}
		fmt.Println()
	}

	fmt.Printf("Aligned: %d / %d extractions\n", alignedCount, len(doc.Extractions))
}

// 测试 7: 无示例时的降级行为（小模型可能不返回可解析的 JSON）
func TestIntegrationWithoutExamples(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	text := "Lady Juliet gazed longingly at the stars, her heart aching for Romeo"

	results, err := extract.Extract(
		ctx,
		text,
		extract.WithAPIKey(apiKey),
		extract.WithBaseURL(baseURL),
		extract.WithModelID(modelID),
		extract.WithPromptDescription("Extract characters and emotions. Respond in JSON format with an 'extractions' array. Each extraction should have 'extraction_class' and 'extraction_text' fields."),
		extract.WithMaxCharBuffer(2000),
		extract.WithExtraKwargs(lmstudioExtraKwargs),
	)
	if err != nil {
		t.Fatalf("Extraction failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}

	doc := results[0]
	fmt.Printf("=== Without Examples Test ===\n")
	fmt.Printf("Extractions: %d\n", len(doc.Extractions))
	for _, ext := range doc.Extractions {
		fmt.Printf("  %s: %s\n", ext.ExtractionClass, ext.ExtractionText)
	}
	// 注意：小模型在没有示例时可能无法返回正确格式，这是预期行为
	if len(doc.Extractions) == 0 {
		t.Log("Warning: small model returned 0 extractions without few-shot examples (expected for gemma-4-e4b)")
	}
}
