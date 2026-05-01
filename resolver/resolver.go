// Package resolver provides LLM output parsing and text alignment functionality.
package resolver

import (
	"fmt"
	"strings"

	"github.com/aqcool/langextract/core"
	"github.com/aqcool/langextract/prompting"
)

const (
	// FuzzyAlignmentMinThreshold is the minimum threshold for fuzzy alignment.
	FuzzyAlignmentMinThreshold = 0.75
	
	// FuzzyAlignmentMinDensity is the minimum density for fuzzy alignment.
	FuzzyAlignmentMinDensity = 1.0 / 3.0
)

// Resolver resolves LLM text outputs into structured data.
type Resolver struct {
	fenceOutput bool
	formatType  core.FormatType
	formatHandler *prompting.FormatHandler
}

// Option is a functional option for Resolver configuration.
type Option func(*Resolver)

// WithFenceOutput sets whether to expect fenced output.
func WithFenceOutput(fenceOutput bool) Option {
	return func(r *Resolver) {
		r.fenceOutput = fenceOutput
	}
}

// WithFormatType sets the format type.
func WithFormatType(formatType core.FormatType) Option {
	return func(r *Resolver) {
		r.formatType = formatType
	}
}

// NewResolver creates a new Resolver.
func NewResolver(opts ...Option) *Resolver {
	r := &Resolver{
		fenceOutput: true,
		formatType:  core.FormatTypeJSON,
	}

	for _, opt := range opts {
		opt(r)
	}

	r.formatHandler = prompting.NewFormatHandler(r.formatType)
	return r
}

// Resolve parses LLM output text into extractions.
func (r *Resolver) Resolve(inputText string) ([]*core.Extraction, error) {
	text := strings.TrimSpace(inputText)
	if text == "" {
		return nil, nil
	}

	// Remove fence markers if present
	if r.fenceOutput {
		text = r.removeFences(text)
	}

	// Parse using format handler
	extractions, err := r.formatHandler.ParseOutput(text)
	if err != nil {
		return nil, fmt.Errorf("failed to parse output: %w", err)
	}

	return extractions, nil
}

// removeFences removes code fence markers from the text.
func (r *Resolver) removeFences(text string) string {
	// Check for ```json or ```yaml at the start
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return text
	}

	startIdx := 0
	endIdx := len(lines)

	// Remove opening fence
	if strings.HasPrefix(strings.TrimSpace(lines[0]), "```") {
		startIdx = 1
	}

	// Remove closing fence
	if endIdx > startIdx && strings.HasPrefix(strings.TrimSpace(lines[endIdx-1]), "```") {
		endIdx = endIdx - 1
	}

	if startIdx >= endIdx {
		return text
	}

	return strings.Join(lines[startIdx:endIdx], "\n")
}

// AlignResult contains the result of alignment.
type AlignResult struct {
	Extraction      *core.Extraction
	AlignmentStatus core.AlignmentStatus
}

// Align aligns extractions with source text, setting token/char intervals and alignment status.
func (r *Resolver) Align(
	extractions []*core.Extraction,
	sourceText string,
	tokenOffset int,
	charOffset *int,
	opts ...AlignOption,
) []*core.Extraction {
	config := &AlignConfig{
		EnableFuzzyAlignment:      true,
		FuzzyAlignmentThreshold:   FuzzyAlignmentMinThreshold,
		AcceptMatchLesser:         true,
		FuzzyAlignmentMinDensity:  FuzzyAlignmentMinDensity,
	}

	for _, opt := range opts {
		opt(config)
	}

	tokenizedSource := core.Tokenize(sourceText)
	result := make([]*core.Extraction, len(extractions))

	for i, extraction := range extractions {
		aligned := r.alignSingle(
			extraction,
			tokenizedSource,
			tokenOffset,
			charOffset,
			config,
		)
		result[i] = aligned
	}

	return result
}

// AlignConfig holds configuration for alignment.
type AlignConfig struct {
	EnableFuzzyAlignment     bool
	FuzzyAlignmentThreshold  float64
	AcceptMatchLesser        bool
	FuzzyAlignmentMinDensity float64
}

// AlignOption is a functional option for alignment configuration.
type AlignOption func(*AlignConfig)

// WithEnableFuzzyAlignment sets whether to enable fuzzy alignment.
func WithEnableFuzzyAlignment(enable bool) AlignOption {
	return func(c *AlignConfig) {
		c.EnableFuzzyAlignment = enable
	}
}

// WithFuzzyAlignmentThreshold sets the fuzzy alignment threshold.
func WithFuzzyAlignmentThreshold(threshold float64) AlignOption {
	return func(c *AlignConfig) {
		c.FuzzyAlignmentThreshold = threshold
	}
}

// WithAcceptMatchLesser sets whether to accept partial exact matches.
func WithAcceptMatchLesser(accept bool) AlignOption {
	return func(c *AlignConfig) {
		c.AcceptMatchLesser = accept
	}
}

// alignSingle aligns a single extraction with the source text.
func (r *Resolver) alignSingle(
	extraction *core.Extraction,
	tokenizedSource *core.TokenizedText,
	tokenOffset int,
	charOffset *int,
	config *AlignConfig,
) *core.Extraction {
	// Clone the extraction
	aligned := extraction.Clone()

	// Try exact match first
	matchResult := r.findExactMatch(extraction.ExtractionText, tokenizedSource)

	if matchResult.Found {
		aligned.AlignmentStatus = core.AlignmentStatusMatchExact
		if matchResult.Partial {
			aligned.AlignmentStatus = core.AlignmentStatusMatchLesser
			if !config.AcceptMatchLesser {
				// Reject partial match
				return aligned
			}
		}

		// Set token interval
		aligned.TokenInterval = &core.TokenInterval{
			StartIndex: matchResult.StartIndex + tokenOffset,
			EndIndex:   matchResult.EndIndex + tokenOffset,
		}

		// Set char interval
		if matchResult.StartChar >= 0 && matchResult.EndChar >= 0 {
			startChar := matchResult.StartChar
			endChar := matchResult.EndChar
			if charOffset != nil {
				startChar += *charOffset
				endChar += *charOffset
			}
			aligned.CharInterval = &core.CharInterval{
				StartPos: &startChar,
				EndPos:   &endChar,
			}
		}

		return aligned
	}

	// Try fuzzy alignment if enabled
	if config.EnableFuzzyAlignment {
		fuzzyResult := r.findFuzzyMatch(
			extraction.ExtractionText,
			tokenizedSource,
			config.FuzzyAlignmentThreshold,
			config.FuzzyAlignmentMinDensity,
		)

		if fuzzyResult.Found {
			aligned.AlignmentStatus = core.AlignmentStatusMatchFuzzy

			// Set token interval
			aligned.TokenInterval = &core.TokenInterval{
				StartIndex: fuzzyResult.StartIndex + tokenOffset,
				EndIndex:   fuzzyResult.EndIndex + tokenOffset,
			}

			// Set char interval
			if fuzzyResult.StartChar >= 0 && fuzzyResult.EndChar >= 0 {
				startChar := fuzzyResult.StartChar
				endChar := fuzzyResult.EndChar
				if charOffset != nil {
					startChar += *charOffset
					endChar += *charOffset
				}
				aligned.CharInterval = &core.CharInterval{
					StartPos: &startChar,
					EndPos:   &endChar,
				}
			}

			return aligned
		}
	}

	// No alignment found
	return aligned
}

// MatchResult represents the result of a text match.
type MatchResult struct {
	Found      bool
	Partial    bool
	StartIndex int
	EndIndex   int
	StartChar  int
	EndChar    int
}

// findExactMatch finds an exact match for the extraction text in the source.
func (r *Resolver) findExactMatch(extractionText string, tokenizedSource *core.TokenizedText) MatchResult {
	// Tokenize the extraction text
	tokenizedExtraction := core.Tokenize(extractionText)
	if len(tokenizedExtraction.Tokens) == 0 {
		return MatchResult{Found: false}
	}

	// Try to find the extraction tokens in the source tokens
	extractionTokens := tokenizedExtraction.Tokens
	sourceTokens := tokenizedSource.Tokens

	for i := 0; i <= len(sourceTokens)-len(extractionTokens); i++ {
		match := true
		for j := 0; j < len(extractionTokens); j++ {
			sourceToken := sourceTokens[i+j]
			extractionToken := extractionTokens[j]
			
			sourceText := getTokenText(tokenizedSource, sourceToken)
			extractionText := getTokenText(tokenizedExtraction, extractionToken)
			
			if !strings.EqualFold(sourceText, extractionText) {
				match = false
				break
			}
		}

		if match {
			// Found exact match
			startToken := sourceTokens[i]
			endToken := sourceTokens[i+len(extractionTokens)-1]

			startChar := 0
			endChar := 0
			if startToken.CharInterval.StartPos != nil {
				startChar = *startToken.CharInterval.StartPos
			}
			if endToken.CharInterval.EndPos != nil {
				endChar = *endToken.CharInterval.EndPos
			}

			return MatchResult{
				Found:      true,
				Partial:    false,
				StartIndex: i,
				EndIndex:   i + len(extractionTokens),
				StartChar:  startChar,
				EndChar:    endChar,
			}
		}
	}

	// Try partial match (extraction longer than match)
	if len(extractionTokens) > 1 {
		for i := 0; i <= len(sourceTokens)-1; i++ {
			sourceToken := sourceTokens[i]
			sourceText := getTokenText(tokenizedSource, sourceToken)
			
			// Check if source token is in extraction
			for _, extractionToken := range extractionTokens {
				extractionText := getTokenText(tokenizedExtraction, extractionToken)
				if strings.EqualFold(sourceText, extractionText) {
					startChar := 0
					endChar := 0
					if sourceToken.CharInterval.StartPos != nil {
						startChar = *sourceToken.CharInterval.StartPos
					}
					if sourceToken.CharInterval.EndPos != nil {
						endChar = *sourceToken.CharInterval.EndPos
					}

					return MatchResult{
						Found:      true,
						Partial:    true,
						StartIndex: i,
						EndIndex:   i + 1,
						StartChar:  startChar,
						EndChar:    endChar,
					}
				}
			}
		}
	}

	return MatchResult{Found: false}
}

// findFuzzyMatch finds a fuzzy match using LCS algorithm.
func (r *Resolver) findFuzzyMatch(
	extractionText string,
	tokenizedSource *core.TokenizedText,
	threshold float64,
	minDensity float64,
) MatchResult {
	// Tokenize the extraction text
	tokenizedExtraction := core.Tokenize(extractionText)
	if len(tokenizedExtraction.Tokens) == 0 {
		return MatchResult{Found: false}
	}

	// Use LCS to find best match
	lcsResult := r.findLCSMatch(tokenizedExtraction, tokenizedSource, threshold, minDensity)

	if lcsResult.Found {
		return lcsResult
	}

	return MatchResult{Found: false}
}

// LCSResult represents the result of LCS matching.
type LCSResult struct {
	Found      bool
	StartIndex int
	EndIndex   int
	StartChar  int
	EndChar    int
	Matches    int
}

// findLCSMatch finds the best match using Longest Common Subsequence.
func (r *Resolver) findLCSMatch(
	tokenizedExtraction *core.TokenizedText,
	tokenizedSource *core.TokenizedText,
	threshold float64,
	minDensity float64,
) MatchResult {
	extractionTokens := tokenizedExtraction.Tokens
	sourceTokens := tokenizedSource.Tokens

	if len(extractionTokens) == 0 || len(sourceTokens) == 0 {
		return MatchResult{Found: false}
	}

	// Build token text arrays for comparison
	extractionTexts := make([]string, len(extractionTokens))
	for i, token := range extractionTokens {
		extractionTexts[i] = getTokenText(tokenizedExtraction, token)
	}

	sourceTexts := make([]string, len(sourceTokens))
	for i, token := range sourceTokens {
		sourceTexts[i] = getTokenText(tokenizedSource, token)
	}

	// Find all matches
	matches := make([][]int, len(extractionTexts))
	for i, extText := range extractionTexts {
		matches[i] = []int{}
		for j, srcText := range sourceTexts {
			if strings.EqualFold(extText, srcText) {
				matches[i] = append(matches[i], j)
			}
		}
	}

	// Find best span using LCS
	bestResult := LCSResult{Found: false}
	minMatches := int(float64(len(extractionTokens)) * threshold)

	for startIdx := 0; startIdx < len(sourceTokens); startIdx++ {
		for endIdx := startIdx + 1; endIdx <= len(sourceTokens); endIdx++ {
			// Count matches in this span
			matchCount := 0
			for _, matchIndices := range matches {
				for _, matchIdx := range matchIndices {
					if matchIdx >= startIdx && matchIdx < endIdx {
						matchCount++
						break
					}
				}
			}

			// Check if this span meets threshold
			if matchCount >= minMatches {
				// Check density
				spanLen := endIdx - startIdx
				density := float64(matchCount) / float64(spanLen)
				
				if density >= minDensity {
					// This is a valid match
					if !bestResult.Found || matchCount > bestResult.Matches {
						startToken := sourceTokens[startIdx]
						endToken := sourceTokens[endIdx-1]

						startChar := 0
						endChar := 0
						if startToken.CharInterval.StartPos != nil {
							startChar = *startToken.CharInterval.StartPos
						}
						if endToken.CharInterval.EndPos != nil {
							endChar = *endToken.CharInterval.EndPos
						}

						bestResult = LCSResult{
							Found:      true,
							StartIndex: startIdx,
							EndIndex:   endIdx,
							StartChar:  startChar,
							EndChar:    endChar,
							Matches:    matchCount,
						}
					}
				}
			}
		}
	}

	return MatchResult{
		Found:      bestResult.Found,
		StartIndex: bestResult.StartIndex,
		EndIndex:   bestResult.EndIndex,
		StartChar:  bestResult.StartChar,
		EndChar:    bestResult.EndChar,
	}
}

// getTokenText extracts the text for a given token.
func getTokenText(tokenizedText *core.TokenizedText, token *core.Token) string {
	if token.CharInterval.StartPos == nil || token.CharInterval.EndPos == nil {
		return ""
	}
	return tokenizedText.Text[*token.CharInterval.StartPos:*token.CharInterval.EndPos]
}
