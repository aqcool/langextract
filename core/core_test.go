package core

import (
	"encoding/json"
	"testing"
)

// ==================== Extraction Tests ====================

func TestExtractionClone(t *testing.T) {
	start, end := 10, 20
	idx := 5
	original := &Extraction{
		ExtractionClass: "person",
		ExtractionText:  "Juliet",
		CharInterval:    &CharInterval{StartPos: &start, EndPos: &end},
		AlignmentStatus: AlignmentStatusMatchExact,
		ExtractionIndex: &idx,
		Attributes:      map[string]interface{}{"mood": "longing"},
	}

	clone := original.Clone()

	// Verify values
	if clone.ExtractionClass != original.ExtractionClass {
		t.Error("ExtractionClass mismatch")
	}
	if clone.ExtractionText != original.ExtractionText {
		t.Error("ExtractionText mismatch")
	}
	if *clone.CharInterval.StartPos != *original.CharInterval.StartPos {
		t.Error("CharInterval.StartPos mismatch")
	}
	if clone.Attributes["mood"] != "longing" {
		t.Error("Attributes mismatch")
	}

	// Verify deep copy - modifying clone shouldn't affect original
	*clone.CharInterval.StartPos = 99
	if *original.CharInterval.StartPos == 99 {
		t.Error("Clone is not deep - CharInterval shared")
	}
	clone.Attributes["mood"] = "happy"
	if original.Attributes["mood"] == "happy" {
		t.Error("Clone is not deep - Attributes shared")
	}
}

func TestExtractionCloneNil(t *testing.T) {
	var e *Extraction
	if e.Clone() != nil {
		t.Error("Clone of nil should return nil")
	}
}

func TestExtractionJSONRoundTrip(t *testing.T) {
	start, end := 0, 6
	ext := &Extraction{
		ExtractionClass: "character",
		ExtractionText:  "ROMEO",
		CharInterval:    &CharInterval{StartPos: &start, EndPos: &end},
		Attributes:      map[string]interface{}{"emotional_state": "wonder"},
	}

	data, err := ext.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var restored Extraction
	if err := restored.FromJSON(data); err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if restored.ExtractionClass != ext.ExtractionClass {
		t.Error("JSON round trip failed for ExtractionClass")
	}
	if restored.ExtractionText != ext.ExtractionText {
		t.Error("JSON round trip failed for ExtractionText")
	}
}

// ==================== Document Tests ====================

func TestNewDocumentAutoID(t *testing.T) {
	doc := NewDocument("test text")
	if doc.DocumentID == "" {
		t.Error("DocumentID should be auto-generated")
	}
	if doc.Text != "test text" {
		t.Error("Text mismatch")
	}
}

func TestNewDocumentWithID(t *testing.T) {
	doc := NewDocument("test", WithDocumentID("my-id"))
	if doc.DocumentID != "my-id" {
		t.Errorf("Expected 'my-id', got '%s'", doc.DocumentID)
	}
}

func TestDocumentClone(t *testing.T) {
	ctx := "extra context"
	doc := NewDocument("hello world", WithDocumentID("doc1"), WithAdditionalContext(ctx))

	clone := doc.Clone()
	if clone.Text != doc.Text {
		t.Error("Text mismatch")
	}

	// Deep copy check
	clone.Text = "changed"
	if doc.Text == "changed" {
		t.Error("Clone is not deep - Text shared")
	}
}

func TestDocumentJSONRoundTrip(t *testing.T) {
	doc := NewDocument("test text", WithDocumentID("doc-1"))
	data, err := doc.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var restored Document
	if err := restored.FromJSON(data); err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if restored.DocumentID != doc.DocumentID {
		t.Error("JSON round trip failed for DocumentID")
	}
}

// ==================== AnnotatedDocument Tests ====================

func TestNewAnnotatedDocument(t *testing.T) {
	start, end := 0, 6
	ad := NewAnnotatedDocument(
		WithAnnotatedDocumentID("test-doc"),
		WithAnnotatedDocumentText("ROMEO speaks"),
		WithExtractions([]*Extraction{
			{ExtractionClass: "character", ExtractionText: "ROMEO",
				CharInterval: &CharInterval{StartPos: &start, EndPos: &end}},
		}),
	)

	if ad.DocumentID != "test-doc" {
		t.Errorf("Expected 'test-doc', got '%s'", ad.DocumentID)
	}
	if ad.Text == nil || *ad.Text != "ROMEO speaks" {
		t.Error("Text mismatch")
	}
	if len(ad.Extractions) != 1 {
		t.Errorf("Expected 1 extraction, got %d", len(ad.Extractions))
	}
}

func TestAnnotatedDocumentClone(t *testing.T) {
	ad := NewAnnotatedDocument(
		WithAnnotatedDocumentID("doc1"),
		WithExtractions([]*Extraction{
			{ExtractionClass: "a", ExtractionText: "b"},
		}),
	)

	clone := ad.Clone()
	clone.Extractions[0].ExtractionText = "changed"
	if ad.Extractions[0].ExtractionText == "changed" {
		t.Error("Clone is not deep - Extractions shared")
	}
}

// ==================== ExampleData Tests ====================

func TestNewExampleData(t *testing.T) {
	ed := NewExampleData("sample text",
		&Extraction{ExtractionClass: "class1", ExtractionText: "text1"},
		&Extraction{ExtractionClass: "class2", ExtractionText: "text2"},
	)

	if ed.Text != "sample text" {
		t.Error("Text mismatch")
	}
	if len(ed.Extractions) != 2 {
		t.Errorf("Expected 2 extractions, got %d", len(ed.Extractions))
	}
}

// ==================== ScoredOutput Tests ====================

func TestScoredOutputString(t *testing.T) {
	score := 0.95
	output := "test output"
	so := ScoredOutput{Score: &score, Output: &output}
	str := so.String()
	if str == "" {
		t.Error("String() should not be empty")
	}
}

func TestScoredOutputNoScore(t *testing.T) {
	output := "test output"
	so := ScoredOutput{Output: &output}
	str := so.String()
	if str == "" {
		t.Error("String() should not be empty")
	}
}

// ==================== Serialization Tests ====================

func TestExtractionsKeyConstant(t *testing.T) {
	if ExtractionsKey != "extractions" {
		t.Errorf("Expected 'extractions', got '%s'", ExtractionsKey)
	}
}

func TestAttributeSuffixConstant(t *testing.T) {
	if AttributeSuffix != "_attributes" {
		t.Errorf("Expected '_attributes', got '%s'", AttributeSuffix)
	}
}

func TestAlignmentStatusConstants(t *testing.T) {
	statuses := []AlignmentStatus{
		AlignmentStatusMatchExact,
		AlignmentStatusMatchGreater,
		AlignmentStatusMatchLesser,
		AlignmentStatusMatchFuzzy,
	}
	expected := []string{"match_exact", "match_greater", "match_lesser", "match_fuzzy"}
	for i, s := range statuses {
		if string(s) != expected[i] {
			t.Errorf("Status %d: expected '%s', got '%s'", i, expected[i], s)
		}
	}
}

func TestExtractionJSONSerialization(t *testing.T) {
	ext := &Extraction{
		ExtractionClass: "person",
		ExtractionText:  "Juliet",
		Attributes:      map[string]interface{}{"age": "young"},
	}

	data, err := json.Marshal(ext)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if result["extraction_class"] != "person" {
		t.Error("JSON field extraction_class mismatch")
	}
	if result["extraction_text"] != "Juliet" {
		t.Error("JSON field extraction_text mismatch")
	}
}
