package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aqcool/langextract/core"
	"github.com/aqcool/langextract/extract"
)

func main() {
	// Set API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Example 1: Basic extraction
	fmt.Println("=== Example 1: Basic Extraction ===")
	basicExample(apiKey)

	// Example 2: Extraction with examples
	fmt.Println("\n=== Example 2: Extraction with Examples ===")
	exampleWithFewShot(apiKey)

	// Example 3: Custom model configuration
	fmt.Println("\n=== Example 3: Custom Model Configuration ===")
	customModelExample(apiKey)
}

func basicExample(apiKey string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Input text
	text := "Lady Juliet gazed longingly at the stars, her heart aching for Romeo"

	// Extract structured information
	results, err := extract.Extract(
		ctx,
		text,
		extract.WithAPIKey(apiKey),
		extract.WithModelID("gpt-4o-mini"),
		extract.WithPromptDescription("Extract characters, emotions, and relationships. Use exact text for extractions."),
	)
	if err != nil {
		log.Printf("Extraction failed: %v", err)
		return
	}

	// Print results
	for _, doc := range results {
		fmt.Printf("Document ID: %s\n", doc.DocumentID)
		fmt.Printf("Extractions:\n")
		for _, ext := range doc.Extractions {
			fmt.Printf("  - Class: %s, Text: %s\n", ext.ExtractionClass, ext.ExtractionText)
			if ext.CharInterval != nil {
				fmt.Printf("    Position: %v-%v\n", ext.CharInterval.StartPos, ext.CharInterval.EndPos)
			}
		}
	}
}

func exampleWithFewShot(apiKey string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Input text
	text := "ROMEO. But soft! What light through yonder window breaks? It is the east, and Juliet is the sun."

	// Define examples to guide the model
	examples := []*core.ExampleData{
		core.NewExampleData(
			"ROMEO. But soft! What light through yonder window breaks?",
			&core.Extraction{
				ExtractionClass: "character",
				ExtractionText:  "ROMEO",
				Attributes:       map[string]interface{}{"emotional_state": "wonder"},
			},
			&core.Extraction{
				ExtractionClass: "emotion",
				ExtractionText:  "But soft!",
				Attributes:       map[string]interface{}{"feeling": "gentle awe"},
			},
		),
	}

	// Extract with examples
	results, err := extract.Extract(
		ctx,
		text,
		extract.WithAPIKey(apiKey),
		extract.WithModelID("gpt-4o-mini"),
		extract.WithPromptDescription("Extract characters, emotions, and relationships in order of appearance. Use exact text for extractions."),
		extract.WithExamples(examples),
	)
	if err != nil {
		log.Printf("Extraction failed: %v", err)
		return
	}

	// Print results as JSON
	jsonData, _ := json.MarshalIndent(results, "", "  ")
	fmt.Println(string(jsonData))
}

func customModelExample(apiKey string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Input text
	text := "The patient was prescribed Aspirin 100mg daily for cardiovascular health."

	// Extract with custom configuration
	results, err := extract.Extract(
		ctx,
		text,
		extract.WithAPIKey(apiKey),
		extract.WithModelID("gpt-4o"),
		extract.WithPromptDescription("Extract medication names, dosages, and frequencies."),
		extract.WithTemperature(0.3),
		extract.WithMaxCharBuffer(500),
		extract.WithEnableFuzzyAlign(true),
	)
	if err != nil {
		log.Printf("Extraction failed: %v", err)
		return
	}

	// Print results
	for _, doc := range results {
		fmt.Printf("Document ID: %s\n", doc.DocumentID)
		for _, ext := range doc.Extractions {
			fmt.Printf("  %s: %s\n", ext.ExtractionClass, ext.ExtractionText)
		}
	}
}
