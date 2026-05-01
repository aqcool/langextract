package core

import (
	"testing"
)

func TestRegexTokenizerBasic(t *testing.T) {
	tokenizer := NewRegexTokenizer()
	result := tokenizer.Tokenize("Hello world")

	if len(result.Tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(result.Tokens))
	}

	if result.Tokens[0].TokenType != TokenTypeWord {
		t.Errorf("expected TokenTypeWord for first token, got %d", result.Tokens[0].TokenType)
	}
}

func TestTokenizerWithNumbers(t *testing.T) {
	result := Tokenize("Patient took 100mg daily")
	if len(result.Tokens) < 3 {
		t.Fatalf("expected at least 3 tokens, got %d", len(result.Tokens))
	}

	// Find the number token
	found := false
	for _, tok := range result.Tokens {
		if tok.TokenType == TokenTypeNumber {
			text := result.Text[*tok.CharInterval.StartPos:*tok.CharInterval.EndPos]
			if text == "100" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected to find number token '100'")
	}
}

func TestTokenizerWithPunctuation(t *testing.T) {
	result := Tokenize("Hello, world!")
	if len(result.Tokens) < 3 {
		t.Fatalf("expected at least 3 tokens, got %d", len(result.Tokens))
	}

	punctCount := 0
	for _, tok := range result.Tokens {
		if tok.TokenType == TokenTypePunctuation {
			punctCount++
		}
	}
	if punctCount < 2 {
		t.Errorf("expected at least 2 punctuation tokens, got %d", punctCount)
	}
}

func TestTokenizerNewlineDetection(t *testing.T) {
	result := Tokenize("First line\nSecond line")
	if len(result.Tokens) < 3 {
		t.Fatalf("expected at least 3 tokens, got %d", len(result.Tokens))
	}

	// "Second" should be first token after newline
	foundAfterNewline := false
	for _, tok := range result.Tokens {
		if tok.FirstTokenAfterNewline {
			text := result.Text[*tok.CharInterval.StartPos:*tok.CharInterval.EndPos]
			if text == "Second" {
				foundAfterNewline = true
			}
		}
	}
	if !foundAfterNewline {
		t.Error("expected 'Second' to be marked as first token after newline")
	}
}

func TestTokenizerEmptyText(t *testing.T) {
	result := Tokenize("")
	if len(result.Tokens) != 0 {
		t.Errorf("expected 0 tokens for empty text, got %d", len(result.Tokens))
	}
}

func TestTokensText(t *testing.T) {
	text := "Hello beautiful world"
	tokenized := Tokenize(text)

	// Get "beautiful" (token index 1)
	ti := &TokenInterval{StartIndex: 1, EndIndex: 2}
	extracted, err := TokensText(tokenized, ti)
	if err != nil {
		t.Fatalf("TokensText failed: %v", err)
	}
	if extracted != "beautiful" {
		t.Errorf("expected 'beautiful', got '%s'", extracted)
	}
}

func TestTokensTextInvalidRange(t *testing.T) {
	tokenized := Tokenize("Hello world")
	ti := &TokenInterval{StartIndex: -1, EndIndex: 2}
	_, err := TokensText(tokenized, ti)
	if err == nil {
		t.Error("expected error for negative start index")
	}
}

func TestFindSentenceRange(t *testing.T) {
	text := "Dr. Smith went home. Then he slept."
	tokens := Tokenize(text)

	sentenceRange, err := FindSentenceRange(text, tokens.Tokens, 0, nil)
	if err != nil {
		t.Fatalf("FindSentenceRange failed: %v", err)
	}
	if sentenceRange == nil {
		t.Fatal("expected non-nil sentence range")
	}

	if sentenceRange.StartIndex != 0 {
		t.Errorf("expected start index 0, got %d", sentenceRange.StartIndex)
	}
}

func TestFindSentenceRangeOutOfBounds(t *testing.T) {
	text := "Hello world."
	tokens := Tokenize(text)

	_, err := FindSentenceRange(text, tokens.Tokens, 100, nil)
	if err == nil {
		t.Error("expected error for out of bounds start index")
	}
}
