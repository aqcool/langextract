package resolver

import (
	"testing"

	"github.com/aqcool/langextract/core"
)

func TestResolverResolveJSON(t *testing.T) {
	r := NewResolver(WithFormatType(core.FormatTypeJSON))

	jsonOutput := `{
		"extractions": [
			{"extraction_class": "person", "extraction_text": "Juliet"},
			{"extraction_class": "emotion", "extraction_text": "longing"}
		]
	}`

	extractions, err := r.Resolve(jsonOutput)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if len(extractions) != 2 {
		t.Fatalf("expected 2 extractions, got %d", len(extractions))
	}
	if extractions[0].ExtractionClass != "person" {
		t.Errorf("expected 'person', got '%s'", extractions[0].ExtractionClass)
	}
}

func TestResolverResolveWithFences(t *testing.T) {
	r := NewResolver(WithFormatType(core.FormatTypeJSON), WithFenceOutput(true))

	fencedOutput := "```json\n" + `{
		"extractions": [
			{"extraction_class": "person", "extraction_text": "Juliet"}
		]
	}` + "\n```"

	extractions, err := r.Resolve(fencedOutput)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if len(extractions) != 1 {
		t.Fatalf("expected 1 extraction, got %d", len(extractions))
	}
	if extractions[0].ExtractionText != "Juliet" {
		t.Errorf("expected 'Juliet', got '%s'", extractions[0].ExtractionText)
	}
}

func TestResolverResolveEmpty(t *testing.T) {
	r := NewResolver(WithFormatType(core.FormatTypeJSON))

	extractions, err := r.Resolve("")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if extractions != nil {
		t.Error("expected nil for empty input")
	}
}

func TestResolverResolveYAML(t *testing.T) {
	r := NewResolver(WithFormatType(core.FormatTypeYAML))

	yamlOutput := `extractions:
  - extraction_class: person
    extraction_text: Juliet
`

	extractions, err := r.Resolve(yamlOutput)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if len(extractions) != 1 {
		t.Fatalf("expected 1 extraction, got %d", len(extractions))
	}
}

func TestResolverAlignExactMatch(t *testing.T) {
	r := NewResolver(WithFormatType(core.FormatTypeJSON))

	sourceText := "Lady Juliet gazed longingly at the stars"
	extractions := []*core.Extraction{
		{ExtractionClass: "character", ExtractionText: "Juliet"},
		{ExtractionClass: "emotion", ExtractionText: "longingly"},
	}

	aligned := r.Align(extractions, sourceText, 0, nil)

	if len(aligned) != 2 {
		t.Fatalf("expected 2 aligned extractions, got %d", len(aligned))
	}

	// Juliet should be found
	if aligned[0].AlignmentStatus != core.AlignmentStatusMatchExact &&
		aligned[0].AlignmentStatus != core.AlignmentStatusMatchLesser {
		t.Errorf("expected match for Juliet, got '%s'", aligned[0].AlignmentStatus)
	}

	// longingly should be found
	if aligned[1].AlignmentStatus != core.AlignmentStatusMatchExact &&
		aligned[1].AlignmentStatus != core.AlignmentStatusMatchLesser {
		t.Errorf("expected match for longingly, got '%s'", aligned[1].AlignmentStatus)
	}

	// Check char interval
	if aligned[1].CharInterval != nil {
		matchedText := sourceText[*aligned[1].CharInterval.StartPos:*aligned[1].CharInterval.EndPos]
		if matchedText != "longingly" {
			t.Errorf("expected 'longingly' in char interval, got '%s'", matchedText)
		}
	}
}

func TestResolverAlignFuzzyMatch(t *testing.T) {
	r := NewResolver(WithFormatType(core.FormatTypeJSON))

	sourceText := "Lady Juliet gazed longingly at the stars, her heart aching for Romeo"
	extractions := []*core.Extraction{
		{ExtractionClass: "character", ExtractionText: "Lady Juliet"}, // exact match
	}

	aligned := r.Align(extractions, sourceText, 0, nil,
		WithEnableFuzzyAlignment(true),
		WithFuzzyAlignmentThreshold(0.5),
	)

	if len(aligned) != 1 {
		t.Fatalf("expected 1 aligned extraction, got %d", len(aligned))
	}

	// Should find a match (exact or fuzzy)
	if aligned[0].CharInterval == nil {
		t.Error("expected char interval for 'Lady Juliet'")
	}
}

func TestResolverAlignWithOffset(t *testing.T) {
	r := NewResolver(WithFormatType(core.FormatTypeJSON))

	sourceText := "First part. Juliet gazed at stars."
	extractions := []*core.Extraction{
		{ExtractionClass: "character", ExtractionText: "Juliet"},
	}

	offset := 12 // "Juliet" starts at char 12 in the full text
	aligned := r.Align(extractions, sourceText, 0, &offset)

	if len(aligned) != 1 {
		t.Fatalf("expected 1 aligned extraction, got %d", len(aligned))
	}

	// Char interval should be offset
	if aligned[0].CharInterval != nil && aligned[0].CharInterval.StartPos != nil {
		if *aligned[0].CharInterval.StartPos < offset {
			t.Errorf("char start should be >= offset %d, got %d", offset, *aligned[0].CharInterval.StartPos)
		}
	}
}

func TestRemoveFences(t *testing.T) {
	r := NewResolver()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "json fence",
			input: "```json\n{\"key\": \"value\"}\n```",
			want:  "{\"key\": \"value\"}",
		},
		{
			name:  "yaml fence",
			input: "```yaml\nkey: value\n```",
			want:  "key: value",
		},
		{
			name:  "no fence",
			input: "{\"key\": \"value\"}",
			want:  "{\"key\": \"value\"}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.removeFences(tt.input)
			if result != tt.want {
				t.Errorf("got '%s', want '%s'", result, tt.want)
			}
		})
	}
}

func TestFindExactMatch(t *testing.T) {
	r := NewResolver()
	tokenizedSource := core.Tokenize("Lady Juliet gazed at the stars")

	result := r.findExactMatch("Juliet", tokenizedSource)
	if !result.Found {
		t.Error("expected to find 'Juliet'")
	}
	if result.Partial {
		t.Error("should not be partial match")
	}
}

func TestFindExactMatchCaseInsensitive(t *testing.T) {
	r := NewResolver()
	tokenizedSource := core.Tokenize("HELLO world")

	result := r.findExactMatch("hello", tokenizedSource)
	if !result.Found {
		t.Error("expected case-insensitive match for 'hello'")
	}
}

func TestFindExactMatchNotFound(t *testing.T) {
	r := NewResolver()
	tokenizedSource := core.Tokenize("Hello world")

	result := r.findExactMatch("xyz", tokenizedSource)
	if result.Found {
		t.Error("should not find 'xyz'")
	}
}

func TestFindExactMatchPartial(t *testing.T) {
	r := NewResolver()
	tokenizedSource := core.Tokenize("The quick brown fox")

	// Search for multi-word that partially matches
	result := r.findExactMatch("quick brown fox", tokenizedSource)
	if !result.Found {
		t.Error("expected to find match for 'quick brown fox'")
	}
}
