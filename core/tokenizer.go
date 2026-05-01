package core

import (
	"regexp"
	"strings"
	"unicode"
)

// TokenType represents the type of a token.
type TokenType int

const (
	TokenTypeWord       TokenType = 0
	TokenTypeNumber     TokenType = 1
	TokenTypePunctuation TokenType = 2
)

// Token represents a token extracted from text.
// Each token is assigned an index and classified into a type (word, number,
// punctuation). The token also records the range of characters (CharInterval)
// that correspond to the substring from the original text.
// Additionally, it tracks whether it follows a newline.
type Token struct {
	Index                 int          `json:"index"`
	TokenType             TokenType    `json:"token_type"`
	CharInterval          *CharInterval `json:"char_interval"`
	FirstTokenAfterNewline bool         `json:"first_token_after_newline"`
}

// TokenizedText holds the result of tokenizing a text string.
type TokenizedText struct {
	Text   string   `json:"text"`
	Tokens []*Token `json:"tokens"`
}

// Tokenizer is the interface for tokenizers.
type Tokenizer interface {
	Tokenize(text string) *TokenizedText
}

// RegexTokenizer is a regex-based tokenizer (default).
// The RegexTokenizer is faster than UnicodeTokenizer for English text because it
// skips involved Unicode handling.
type RegexTokenizer struct{}

// NewRegexTokenizer creates a new RegexTokenizer.
func NewRegexTokenizer() *RegexTokenizer {
	return &RegexTokenizer{}
}

// Tokenize splits text into tokens (words, digits, or punctuation).
// Each token is annotated with its character position and type. Tokens
// following a newline or carriage return have FirstTokenAfterNewline set to true.
func (t *RegexTokenizer) Tokenize(text string) *TokenizedText {
	tokenized := &TokenizedText{
		Text:   text,
		Tokens: []*Token{},
	}

	// Pattern for letters (Unicode-aware)
	lettersPattern := `[^\W\d_]+`
	// Pattern for digits
	digitsPattern := `\d+`
	// Pattern for symbols (consecutive non-word non-space characters)
	// Note: Go's RE2 engine doesn't support backreferences (\1),
	// so we use a simpler pattern that groups consecutive symbols.
	symbolsPattern := `[^\w\s]+`

	// Combined pattern
	tokenPattern := regexp.MustCompile(lettersPattern + `|` + digitsPattern + `|` + symbolsPattern)
	
	// Word pattern for classification
	wordPattern := regexp.MustCompile(`^(?:` + lettersPattern + `|` + digitsPattern + `)$`)
	
	// Digits pattern for classification
	digitsOnlyPattern := regexp.MustCompile(`^\d+$`)

	matches := tokenPattern.FindAllStringIndex(text, -1)
	previousEnd := 0

	for tokenIndex, match := range matches {
		startPos, endPos := match[0], match[1]
		matchedText := text[startPos:endPos]

		token := &Token{
			Index:        tokenIndex,
			CharInterval: &CharInterval{StartPos: &startPos, EndPos: &endPos},
			TokenType:    TokenTypeWord,
			FirstTokenAfterNewline: false,
		}

		// Check for newline in gap between previous token and this one
		if tokenIndex > 0 && startPos > previousEnd {
			gap := text[previousEnd:startPos]
			if strings.Contains(gap, "\n") || strings.Contains(gap, "\r") {
				token.FirstTokenAfterNewline = true
			}
		}

		// Determine token type
		if digitsOnlyPattern.MatchString(matchedText) {
			token.TokenType = TokenTypeNumber
		} else if wordPattern.MatchString(matchedText) {
			token.TokenType = TokenTypeWord
		} else {
			token.TokenType = TokenTypePunctuation
		}

		tokenized.Tokens = append(tokenized.Tokens, token)
		previousEnd = endPos
	}

	return tokenized
}

// Default tokenizer instance
var defaultTokenizer = NewRegexTokenizer()

// Tokenize splits text into tokens using the default tokenizer (RegexTokenizer).
func Tokenize(text string) *TokenizedText {
	return defaultTokenizer.Tokenize(text)
}

// TokenizeWithTokenizer splits text into tokens using a custom tokenizer.
func TokenizeWithTokenizer(text string, tokenizer Tokenizer) *TokenizedText {
	return tokenizer.Tokenize(text)
}

// TokensText reconstructs the substring of the original text spanning a given token interval.
func TokensText(tokenizedText *TokenizedText, tokenInterval *TokenInterval) (string, error) {
	if tokenInterval.StartIndex == tokenInterval.EndIndex {
		return "", nil
	}

	if tokenInterval.StartIndex < 0 ||
		tokenInterval.EndIndex > len(tokenizedText.Tokens) ||
		tokenInterval.StartIndex > tokenInterval.EndIndex {
		return "", NewInvalidTokenIntervalError(
			"Invalid token interval. start_index=%d, end_index=%d, total_tokens=%d",
			tokenInterval.StartIndex,
			tokenInterval.EndIndex,
			len(tokenizedText.Tokens),
		)
	}

	startToken := tokenizedText.Tokens[tokenInterval.StartIndex]
	endToken := tokenizedText.Tokens[tokenInterval.EndIndex-1]

	if startToken.CharInterval.StartPos == nil || endToken.CharInterval.EndPos == nil {
		return "", NewInternalError("token char interval is nil")
	}

	return tokenizedText.Text[*startToken.CharInterval.StartPos:*endToken.CharInterval.EndPos], nil
}

// FindSentenceRange finds a 'sentence' interval from a given start index.
// Sentence boundaries are defined by:
// - punctuation tokens ending with sentence markers (.?!)
// - newline breaks followed by an uppercase letter
// - not abbreviations in known abbreviations (e.g., "Dr.")
func FindSentenceRange(text string, tokens []*Token, startTokenIndex int, knownAbbreviations []string) (*TokenInterval, error) {
	if len(tokens) == 0 {
		return &TokenInterval{StartIndex: 0, EndIndex: 0}, nil
	}

	if startTokenIndex < 0 || startTokenIndex >= len(tokens) {
		return nil, NewSentenceRangeError(
			"start_token_index=%d out of range. Total tokens: %d",
			startTokenIndex,
			len(tokens),
		)
	}

	// Default known abbreviations
	if len(knownAbbreviations) == 0 {
		knownAbbreviations = []string{"Mr.", "Mrs.", "Ms.", "Dr.", "Prof.", "St."}
	}

	abbrevSet := make(map[string]bool)
	for _, abbrev := range knownAbbreviations {
		abbrevSet[abbrev] = true
	}

	// End of sentence pattern
	endOfSentencePattern := regexp.MustCompile(`[.?!。！？]["'"»)\]}]*$`)
	closingPunctuation := []string{`"`, `'`, `"`, `'`, "»", ")", "]", "}"}

	i := startTokenIndex
	for i < len(tokens) {
		if tokens[i].TokenType == TokenTypePunctuation {
			tokenText := getTokenText(text, tokens[i])
			if endOfSentencePattern.MatchString(tokenText) {
				// Check if it's an abbreviation
				if i > 0 {
					prevTokenText := getTokenText(text, tokens[i-1])
					combined := prevTokenText + tokenText
					if abbrevSet[combined] {
						i++
						continue
					}
				}

				endIndex := i + 1
				// Consume trailing closing punctuation
				for endIndex < len(tokens) {
					nextTokenText := getTokenText(text, tokens[endIndex])
					if tokens[endIndex].TokenType == TokenTypePunctuation &&
						containsString(closingPunctuation, nextTokenText) {
						endIndex++
					} else {
						break
					}
				}
				return &TokenInterval{StartIndex: startTokenIndex, EndIndex: endIndex}, nil
			}
		}

		// Check for sentence break after newline
		if isSentenceBreakAfterNewline(text, tokens, i) {
			return &TokenInterval{StartIndex: startTokenIndex, EndIndex: i + 1}, nil
		}

		i++
	}

	return &TokenInterval{StartIndex: startTokenIndex, EndIndex: len(tokens)}, nil
}

// getTokenText extracts the text for a given token.
func getTokenText(text string, token *Token) string {
	if token.CharInterval.StartPos == nil || token.CharInterval.EndPos == nil {
		return ""
	}
	return text[*token.CharInterval.StartPos:*token.CharInterval.EndPos]
}

// isSentenceBreakAfterNewline checks if the next token starts uppercase and follows a newline.
func isSentenceBreakAfterNewline(text string, tokens []*Token, currentIdx int) bool {
	if currentIdx+1 >= len(tokens) {
		return false
	}

	nextToken := tokens[currentIdx+1]
	if !nextToken.FirstTokenAfterNewline {
		return false
	}

	nextTokenText := getTokenText(text, nextToken)
	if len(nextTokenText) == 0 {
		return false
	}

	// Assume break unless lowercase (covers numbers/quotes)
	return !unicode.IsLower(rune(nextTokenText[0]))
}

// containsString checks if a string is in a slice.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// InvalidTokenIntervalError is raised when a token interval is invalid or out of range.
type InvalidTokenIntervalError struct {
	*LangExtractError
}

func (e *InvalidTokenIntervalError) Error() string {
	return e.Message
}

func NewInvalidTokenIntervalError(format string, args ...interface{}) *InvalidTokenIntervalError {
	return &InvalidTokenIntervalError{
		LangExtractError: NewLangExtractError(formatArgs(format, args...)),
	}
}

// SentenceRangeError is raised when the start token index for a sentence is out of range.
type SentenceRangeError struct {
	*LangExtractError
}

func (e *SentenceRangeError) Error() string {
	return e.Message
}

func NewSentenceRangeError(format string, args ...interface{}) *SentenceRangeError {
	return &SentenceRangeError{
		LangExtractError: NewLangExtractError(formatArgs(format, args...)),
	}
}

// formatArgs formats a message with arguments.
func formatArgs(format string, args ...interface{}) string {
	if len(args) == 0 {
		return format
	}
	return strings.ReplaceAll(format, "%d", strings.Join([]string{format}, " "))
}