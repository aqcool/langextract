// Package chunking provides text chunking functionality for large documents.
package chunking

import (
	"fmt"
	"strings"

	"github.com/aqcool/langextract/core"
)

// TextChunk stores a text chunk with attributes to the source document.
type TextChunk struct {
	TokenInterval *core.TokenInterval
	Document      *core.Document
	chunkText     *string
	charInterval  *core.CharInterval
}

// NewTextChunk creates a new TextChunk.
func NewTextChunk(tokenInterval *core.TokenInterval, document *core.Document) *TextChunk {
	return &TextChunk{
		TokenInterval: tokenInterval,
		Document:      document,
	}
}

// DocumentID returns the document ID from the source document.
func (tc *TextChunk) DocumentID() string {
	if tc.Document != nil {
		return tc.Document.DocumentID
	}
	return ""
}

// ChunkText returns the chunk text.
func (tc *TextChunk) ChunkText() (string, error) {
	if tc.chunkText != nil {
		return *tc.chunkText, nil
	}

	if tc.Document == nil || tc.Document.Text == "" {
		return "", fmt.Errorf("document text must be set to access chunk_text")
	}

	tokenizedText := core.Tokenize(tc.Document.Text)
	text, err := core.TokensText(tokenizedText, tc.TokenInterval)
	if err != nil {
		return "", err
	}

	tc.chunkText = &text
	return text, nil
}

// SanitizedChunkText returns the sanitized chunk text.
func (tc *TextChunk) SanitizedChunkText() (string, error) {
	text, err := tc.ChunkText()
	if err != nil {
		return "", err
	}
	return sanitize(text), nil
}

// CharInterval returns the character interval corresponding to the token interval.
func (tc *TextChunk) CharInterval() (*core.CharInterval, error) {
	if tc.charInterval != nil {
		return tc.charInterval, nil
	}

	if tc.Document == nil || tc.Document.Text == "" {
		return nil, fmt.Errorf("document text must be set to compute char_interval")
	}

	tokenizedText := core.Tokenize(tc.Document.Text)
	if tc.TokenInterval.StartIndex < 0 || tc.TokenInterval.EndIndex > len(tokenizedText.Tokens) {
		return nil, fmt.Errorf("token interval out of range")
	}

	if tc.TokenInterval.StartIndex >= tc.TokenInterval.EndIndex {
		return nil, fmt.Errorf("invalid token interval")
	}

	startToken := tokenizedText.Tokens[tc.TokenInterval.StartIndex]
	endToken := tokenizedText.Tokens[tc.TokenInterval.EndIndex-1]

	if startToken.CharInterval.StartPos == nil || endToken.CharInterval.EndPos == nil {
		return nil, fmt.Errorf("token char interval is nil")
	}

	tc.charInterval = &core.CharInterval{
		StartPos: startToken.CharInterval.StartPos,
		EndPos:   endToken.CharInterval.EndPos,
	}

	return tc.charInterval, nil
}

// AdditionalContext returns the additional context for prompting from the source document.
func (tc *TextChunk) AdditionalContext() *string {
	if tc.Document != nil {
		return tc.Document.AdditionalContext
	}
	return nil
}

// sanitize removes problematic characters from text.
func sanitize(text string) string {
	// Remove control characters except newlines and tabs
	var result strings.Builder
	for _, r := range text {
		if r == '\n' || r == '\t' || r >= 32 {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// TextChunker breaks documents into chunks of sentences.
type TextChunker struct {
	maxCharBuffer int
	tokenizer     core.Tokenizer
}

// Option is a functional option for TextChunker configuration.
type Option func(*TextChunker)

// WithMaxCharBuffer sets the maximum character buffer size.
func WithMaxCharBuffer(maxChars int) Option {
	return func(tc *TextChunker) {
		tc.maxCharBuffer = maxChars
	}
}

// WithTokenizer sets a custom tokenizer.
func WithTokenizer(tokenizer core.Tokenizer) Option {
	return func(tc *TextChunker) {
		tc.tokenizer = tokenizer
	}
}

// NewTextChunker creates a new TextChunker.
func NewTextChunker(opts ...Option) *TextChunker {
	chunker := &TextChunker{
		maxCharBuffer: 1000,
		tokenizer:     core.NewRegexTokenizer(),
	}

	for _, opt := range opts {
		opt(chunker)
	}

	return chunker
}

// ChunkDocument breaks a document into chunks based on sentence boundaries.
func (tc *TextChunker) ChunkDocument(document *core.Document) ([]*TextChunk, error) {
	if document == nil || document.Text == "" {
		return nil, fmt.Errorf("document text is empty")
	}

	tokenizedText := tc.tokenizer.Tokenize(document.Text)
	if len(tokenizedText.Tokens) ==  {
		return []*TextChunk{}, nil
	}

	var chunks []*TextChunk
	currentStart := 0
	currentLength := 0

	for i := 0; i < len(tokenizedText.Tokens); i++ {
		token := tokenizedText.Tokens[i]
		
		// Get token text length
		if token.CharInterval.StartPos == nil || token.CharInterval.EndPos == nil {
			continue
		}
		tokenLength := *token.CharInterval.EndPos - *token.CharInterval.StartPos

		// Check if we should start a new chunk
		if currentLength+tokenLength > tc.maxCharBuffer && currentStart < i {
			// Find sentence boundary
			sentenceEnd := tc.findSentenceBoundary(tokenizedText, currentStart, i)
			if sentenceEnd > currentStart {
				chunks = append(chunks, NewTextChunk(
					&core.TokenInterval{
						StartIndex: currentStart,
						EndIndex:   sentenceEnd,
					},
					document,
				))
				currentStart = sentenceEnd
				currentLength = 0
				i = currentStart - 1 // Will be incremented in the loop
				continue
			}
		}

		currentLength += tokenLength
	}

	// Add remaining chunk
	if currentStart < len(tokenizedText.Tokens) {
		chunks = append(chunks, NewTextChunk(
			&core.TokenInterval{
				StartIndex: currentStart,
				EndIndex:   len(tokenizedText.Tokens),
			},
			document,
		))
	}

	return chunks, nil
}

// findSentenceBoundary finds a sentence boundary within the given range.
func (tc *TextChunker) findSentenceBoundary(tokenizedText *core.TokenizedText, start, end int) int {
	// Try to find sentence boundary from end backwards
	for i := end - 1; i >= start; i-- {
		token := tokenizedText.Tokens[i]
		if token.TokenType == core.TokenTypePunctuation {
			tokenText := getTokenText(tokenizedText, token)
			if isEndOfSentence(tokenText) {
				return i + 1
			}
		}
	}
	return end
}

// getTokenText extracts the text for a given token.
func getTokenText(tokenizedText *core.TokenizedText, token *core.Token) string {
	if token.CharInterval.StartPos == nil || token.CharInterval.EndPos == nil {
		return ""
	}
	return tokenizedText.Text[*token.CharInterval.StartPos:*token.CharInterval.EndPos]
}

// isEndOfSentence checks if a punctuation token ends a sentence.
func isEndOfSentence(text string) bool {
	if len(text) == 0 {
		return false
	}
	lastChar := text[len(text)-1]
	return lastChar == '.' || lastChar == '!' || lastChar == '?' ||
		lastChar == '。' || lastChar == '！' || lastChar == '？'
}

// ChunkDocuments chunks multiple documents.
func (tc *TextChunker) ChunkDocuments(documents []*core.Document) ([][]*TextChunk, error) {
	result := make([][]*TextChunk, len(documents))
	for i, doc := range documents {
		chunks, err := tc.ChunkDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to chunk document %s: %w", doc.DocumentID, err)
		}
		result[i] = chunks
	}
	return result, nil
}
