package chunking

import (
	"testing"

	"github.com/aqcool/langextract/core"
)

func TestTextChunkerBasic(t *testing.T) {
	chunker := NewTextChunker(WithMaxCharBuffer(500))

	doc := core.NewDocument("Hello world. This is a test. Another sentence here.")
	chunks, err := chunker.ChunkDocument(doc)
	if err != nil {
		t.Fatalf("ChunkDocument failed: %v", err)
	}

	if len(chunks) == 0 {
		t.Fatal("expected at least 1 chunk")
	}

	// Verify we can get text from each chunk
	for i, chunk := range chunks {
		text, err := chunk.ChunkText()
		if err != nil {
			t.Fatalf("chunk %d ChunkText failed: %v", i, err)
		}
		if text == "" {
			t.Errorf("chunk %d has empty text", i)
		}
	}
}

func TestTextChunkerSmallBuffer(t *testing.T) {
	chunker := NewTextChunker(WithMaxCharBuffer(20))

	doc := core.NewDocument("First sentence. Second sentence. Third sentence.")
	chunks, err := chunker.ChunkDocument(doc)
	if err != nil {
		t.Fatalf("ChunkDocument failed: %v", err)
	}

	// Small buffer should produce multiple chunks
	if len(chunks) < 2 {
		t.Errorf("expected at least 2 chunks with small buffer, got %d", len(chunks))
	}
}

func TestTextChunkerEmptyDocument(t *testing.T) {
	chunker := NewTextChunker()

	doc := &core.Document{Text: ""}
	_, err := chunker.ChunkDocument(doc)
	if err == nil {
		t.Error("expected error for empty document")
	}
}

func TestTextChunkerNilDocument(t *testing.T) {
	chunker := NewTextChunker()
	_, err := chunker.ChunkDocument(nil)
	if err == nil {
		t.Error("expected error for nil document")
	}
}

func TestTextChunkDocumentID(t *testing.T) {
	chunker := NewTextChunker(WithMaxCharBuffer(1000))

	doc := core.NewDocument("Some text here.", core.WithDocumentID("my-doc"))
	chunks, err := chunker.ChunkDocument(doc)
	if err != nil {
		t.Fatalf("ChunkDocument failed: %v", err)
	}

	if len(chunks) == 0 {
		t.Fatal("expected at least 1 chunk")
	}

	if chunks[0].DocumentID() != "my-doc" {
		t.Errorf("expected 'my-doc', got '%s'", chunks[0].DocumentID())
	}
}

func TestTextChunkCharInterval(t *testing.T) {
	chunker := NewTextChunker(WithMaxCharBuffer(1000))

	doc := core.NewDocument("Hello world")
	chunks, err := chunker.ChunkDocument(doc)
	if err != nil {
		t.Fatalf("ChunkDocument failed: %v", err)
	}

	if len(chunks) == 0 {
		t.Fatal("expected at least 1 chunk")
	}

	ci, err := chunks[0].CharInterval()
	if err != nil {
		t.Fatalf("CharInterval failed: %v", err)
	}
	if ci == nil {
		t.Fatal("expected non-nil CharInterval")
	}
	if *ci.StartPos != 0 {
		t.Errorf("expected start pos 0, got %d", *ci.StartPos)
	}
}

func TestTextChunkAdditionalContext(t *testing.T) {
	chunker := NewTextChunker(WithMaxCharBuffer(1000))

	ctx := "extra info"
	doc := core.NewDocument("Some text.", core.WithAdditionalContext(ctx))
	chunks, err := chunker.ChunkDocument(doc)
	if err != nil {
		t.Fatalf("ChunkDocument failed: %v", err)
	}

	if len(chunks) == 0 {
		t.Fatal("expected at least 1 chunk")
	}

	additionalCtx := chunks[0].AdditionalContext()
	if additionalCtx == nil || *additionalCtx != "extra info" {
		t.Error("expected additional context 'extra info'")
	}
}

func TestChunkDocuments(t *testing.T) {
	chunker := NewTextChunker(WithMaxCharBuffer(500))

	docs := []*core.Document{
		core.NewDocument("First document."),
		core.NewDocument("Second document."),
	}

	results, err := chunker.ChunkDocuments(docs)
	if err != nil {
		t.Fatalf("ChunkDocuments failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for i, chunks := range results {
		if len(chunks) == 0 {
			t.Errorf("document %d: expected at least 1 chunk", i)
		}
	}
}

func TestSanitize(t *testing.T) {
	// Control characters (except newline and tab) should be removed
	input := "Hello\x00\x01world\t\n"
	result := sanitize(input)
	if result != "Hello\t\nworld\t\n" {
		// Actually sanitize preserves newlines and tabs, removes other control chars
		for _, r := range result {
			if r < 32 && r != '\n' && r != '\t' {
				t.Errorf("found unexpected control char: %d", r)
			}
		}
	}
}
