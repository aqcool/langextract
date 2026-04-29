# LangExtract Go

English | [中文](README_CN.md)

A Go library for extracting structured information from unstructured text using LLMs.

## Overview

LangExtract Go is a Golang implementation of Google's [LangExtract](https://github.com/google/langextract) library. It uses Large Language Models (LLMs) to extract structured information from text documents based on user-defined instructions and examples.

## Features

- **Structured Information Extraction**: Extract structured data from unstructured text using LLMs
- **Precise Source Grounding**: Map every extraction to its exact location in the source text
- **OpenAI API Compatible**: Support for standard OpenAI API (including OpenAI-compatible services)
- **Long Document Processing**: Text chunking, parallel processing, and multi-pass extraction for better recall
- **Structured Output**: JSON/YAML output format with schema constraints
- **Text Alignment**: LCS (Longest Common Subsequence) algorithm for fuzzy text matching

## Installation

```bash
go get github.com/aqcool/langextract
```

## Quick Start

### Basic Extraction

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/aqcool/langextract/core"
    "github.com/aqcool/langextract/extract"
)

func main() {
    ctx := context.Background()
    
    // Input text
    text := "Lady Juliet gazed longingly at the stars, her heart aching for Romeo"
    
    // Extract structured information
    results, err := extract.Extract(
        ctx,
        text,
        extract.WithAPIKey("your-api-key"),
        extract.WithModelID("gpt-4o-mini"),
        extract.WithPromptDescription("Extract characters, emotions, and relationships."),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Print results
    for _, doc := range results {
        fmt.Printf("Document: %s\n", doc.DocumentID)
        for _, ext := range doc.Extractions {
            fmt.Printf("  %s: %s\n", ext.ExtractionClass, ext.ExtractionText)
        }
    }
}
```

### Extraction with Few-Shot Examples

```go
// Define examples to guide the model
examples := []*core.ExampleData{
    core.NewExampleData(
        "ROMEO. But soft! What light through yonder window breaks?",
        &core.Extraction{
            ExtractionClass: "character",
            ExtractionText:  "ROMEO",
            Attributes:       map[string]interface{}{"emotional_state": "wonder"},
        },
    ),
}

results, err := extract.Extract(
    ctx,
    text,
    extract.WithAPIKey("your-api-key"),
    extract.WithPromptDescription("Extract characters and emotions."),
    extract.WithExamples(examples),
)
```

## Configuration Options

### Model Configuration

- `WithModelID(modelID string)`: Set the LLM model ID (default: "gpt-4o-mini")
- `WithAPIKey(apiKey string)`: Set the API key (or use `OPENAI_API_KEY` env var)
- `WithBaseURL(baseURL string)`: Set custom API base URL for OpenAI-compatible services
- `WithTemperature(temp float64)`: Set sampling temperature
- `WithMaxWorkers(workers int)`: Set number of parallel workers (default: 10)

### Extraction Configuration

- `WithPromptDescription(desc string)`: Set extraction instructions (required)
- `WithExamples(examples []*core.ExampleData)`: Set few-shot examples
- `WithMaxCharBuffer(chars int)`: Set maximum chunk size (default: 1000)
- `WithExtractionPasses(passes int)`: Set number of extraction passes (default: 1)
- `WithFormatType(formatType core.FormatType)`: Set output format (JSON or YAML)

### Alignment Configuration

- `WithEnableFuzzyAlign(enable bool)`: Enable/disable fuzzy text alignment
- `WithFuzzyThreshold(threshold float64)`: Set fuzzy alignment threshold (default: 0.75)

## API Reference

### Main API

```go
// Extract structured information from text
func Extract(ctx context.Context, textOrDocuments interface{}, opts ...Option) ([]*core.AnnotatedDocument, error)

// Extract from file
func ExtractFromFile(ctx context.Context, path string, opts ...Option) ([]*core.AnnotatedDocument, error)
```

### Core Types

```go
// Extraction represents a single extracted entity
type Extraction struct {
    ExtractionClass  string
    ExtractionText   string
    CharInterval     *CharInterval
    AlignmentStatus  AlignmentStatus
    Attributes       map[string]interface{}
}

// AnnotatedDocument represents a document with extractions
type AnnotatedDocument struct {
    DocumentID   string
    Text         *string
    Extractions  []*Extraction
}

// ExampleData represents a few-shot example
type ExampleData struct {
    Text        string
    Extractions []*Extraction
}
```

## Architecture

```
langextract/
├── core/           # Core data models and types
│   ├── data.go     # Extraction, Document, AnnotatedDocument
│   ├── types.go    # FormatType, Constraint, ScoredOutput
│   ├── errors.go   # Custom error types
│   └── tokenizer.go # Text tokenization
├── providers/      # LLM provider interfaces
│   ├── base.go     # BaseLanguageModel interface
│   └── openai/     # OpenAI-compatible provider
├── chunking/       # Text chunking
├── prompting/      # Prompt building
├── resolver/       # Output parsing and text alignment
├── factory/        # Model factory
└── extract/        # Main extraction API
```

## Environment Variables

- `OPENAI_API_KEY`: OpenAI API key
- `LANGEXTRACT_API_KEY`: Fallback API key

## Examples

See the `examples/` directory for more usage examples:

- `examples/basic/`: Basic extraction examples
- `examples/advanced/`: Advanced usage with custom configurations

## Comparison with Python Version

This Go implementation maintains API compatibility with the Python version while leveraging Go's strengths:

- **Concurrency**: Native goroutine support for parallel processing
- **Type Safety**: Strong typing with Go's type system
- **Performance**: Compiled binary with no runtime interpretation
- **Deployment**: Single binary deployment

## License

Apache License 2.0

## Contributing

Contributions are welcome! Please read the contributing guidelines before submitting PRs.

## Acknowledgments

This project is a Go implementation of Google's [LangExtract](https://github.com/google/langextract) library.

## Disclaimer

This is not an officially supported Google product. Use is subject to the Apache 2.0 License.
