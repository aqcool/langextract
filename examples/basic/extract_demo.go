package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aqcool/langextract/core"
	"github.com/aqcool/langextract/extract"
	"github.com/aqcool/langextract/factory"
)

const (
	baseURL = "http://192.168.0.154:1234/v1"
	apiKey  = "sk-lm-xYSLUVJw:ct9kkJ6QYfKQORTWbYgH"
	modelID = "google/gemma-4-e4b"
)

func main() {
	// 创建兼容 LMStudio 的模型
	model, err := factory.CreateModel(factory.NewModelConfig(
		modelID,
		factory.WithProvider("openai"),
		factory.WithProviderKwargs(map[string]interface{}{
			"api_key":              apiKey,
			"base_url":             baseURL,
			"format_type":          core.FormatTypeJSON,
			"response_format_type": "text",
		}),
	))
	if err != nil {
		log.Fatalf("创建模型失败: %v", err)
	}

	// 定义 few-shot 示例引导模型输出格式
	examples := []*core.ExampleData{
		core.NewExampleData(
			"ROMEO. But soft! What light through yonder window breaks? It is the east, and Juliet is the sun.",
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
			&core.Extraction{
				ExtractionClass: "metaphor",
				ExtractionText:  "Juliet is the sun",
				Attributes:      map[string]interface{}{"meaning": "Juliet brings light and warmth"},
			},
		),
	}

	text := `Lady Juliet gazed longingly at the stars, her heart aching for Romeo.
She whispered softly to the night, "Where art thou, my love?"
The cold stone walls of the Capulet mansion pressed in around her,
but her thoughts flew free to the one who had stolen her heart.`

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	fmt.Println("=== LangExtract 抽取测试 ===")
	fmt.Printf("模型: %s\n", modelID)
	fmt.Printf("输入文本: %d 字符\n\n", len(text))

	start := time.Now()
	results, err := extract.Extract(
		ctx,
		text,
		extract.WithModel(model),
		extract.WithPromptDescription("Extract characters, emotions, metaphors, and relationships from the text. Use exact text for extractions."),
		extract.WithExamples(examples),
		extract.WithMaxCharBuffer(2000),
		extract.WithEnableFuzzyAlign(true),
	)
	if err != nil {
		log.Fatalf("抽取失败: %v", err)
	}

	elapsed := time.Since(start)
	fmt.Printf("耗时: %v\n\n", elapsed)

	for _, doc := range results {
		fmt.Printf("文档 ID: %s\n", doc.DocumentID)
		fmt.Printf("提取数量: %d\n\n", len(doc.Extractions))

		for i, ext := range doc.Extractions {
			fmt.Printf("[%d] 类别: %s\n", i+1, ext.ExtractionClass)
			fmt.Printf("    文本: %s\n", ext.ExtractionText)

			if ext.Attributes != nil && len(ext.Attributes) > 0 {
				attrJSON, _ := json.Marshal(ext.Attributes)
				fmt.Printf("    属性: %s\n", string(attrJSON))
			}

			if ext.CharInterval != nil && ext.CharInterval.StartPos != nil && ext.CharInterval.EndPos != nil {
				matched := text[*ext.CharInterval.StartPos:*ext.CharInterval.EndPos]
				fmt.Printf("    位置: %d-%d (对齐: %s)", *ext.CharInterval.StartPos, *ext.CharInterval.EndPos, ext.AlignmentStatus)
				if matched != ext.ExtractionText {
					fmt.Printf(" [原文: \"%s\"]", matched)
				}
				fmt.Println()
			}
			fmt.Println()
		}
	}

	// 输出完整 JSON
	fmt.Println("=== 完整 JSON 输出 ===")
	jsonData, _ := json.MarshalIndent(results, "", "  ")
	fmt.Println(string(jsonData))
}
