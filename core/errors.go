package core

import (
	"errors"
	"fmt"
)

// Base error types for LangExtract.
// All errors defined here allow users to catch LangExtract-specific errors.

// LangExtractError is the base error for all LangExtract errors.
type LangExtractError struct {
	Message string
}

func (e *LangExtractError) Error() string {
	return e.Message
}

// NewLangExtractError creates a new LangExtractError.
func NewLangExtractError(message string) *LangExtractError {
	return &LangExtractError{Message: message}
}

// InferenceError is the base error for inference-related errors.
type InferenceError struct {
	*LangExtractError
}

func (e *InferenceError) Error() string {
	return e.Message
}

// NewInferenceError creates a new InferenceError.
func NewInferenceError(message string) *InferenceError {
	return &InferenceError{
		LangExtractError: NewLangExtractError(message),
	}
}

// InferenceConfigError is raised for configuration errors.
// This includes missing API keys, invalid model IDs, or other configuration issues.
type InferenceConfigError struct {
	*InferenceError
}

func (e *InferenceConfigError) Error() string {
	return e.Message
}

// NewInferenceConfigError creates a new InferenceConfigError.
func NewInferenceConfigError(message string) *InferenceConfigError {
	return &InferenceConfigError{
		InferenceError: NewInferenceError(message),
	}
}

// InferenceRuntimeError is raised for runtime inference errors.
// This includes API call failures, network errors, or other runtime issues.
type InferenceRuntimeError struct {
	*InferenceError
	Original error
	Provider string
}

func (e *InferenceRuntimeError) Error() string {
	if e.Original != nil {
		return fmt.Sprintf("%s (provider: %s, original: %v)", e.Message, e.Provider, e.Original)
	}
	return fmt.Sprintf("%s (provider: %s)", e.Message, e.Provider)
}

// Unwrap returns the original error for error chain support.
func (e *InferenceRuntimeError) Unwrap() error {
	return e.Original
}

// NewInferenceRuntimeError creates a new InferenceRuntimeError.
func NewInferenceRuntimeError(message string, original error, provider string) *InferenceRuntimeError {
	return &InferenceRuntimeError{
		InferenceError: NewInferenceError(message),
		Original:       original,
		Provider:       provider,
	}
}

// InferenceOutputError is raised when no scored outputs are available from the language model.
type InferenceOutputError struct {
	*LangExtractError
}

func (e *InferenceOutputError) Error() string {
	return e.Message
}

// NewInferenceOutputError creates a new InferenceOutputError.
func NewInferenceOutputError(message string) *InferenceOutputError {
	return &InferenceOutputError{
		LangExtractError: NewLangExtractError(message),
	}
}

// InvalidDocumentError is raised when document input is invalid.
// This includes cases like duplicate document IDs or malformed documents.
type InvalidDocumentError struct {
	*LangExtractError
}

func (e *InvalidDocumentError) Error() string {
	return e.Message
}

// NewInvalidDocumentError creates a new InvalidDocumentError.
func NewInvalidDocumentError(message string) *InvalidDocumentError {
	return &InvalidDocumentError{
		LangExtractError: NewLangExtractError(message),
	}
}

// InternalError is raised for internal invariant violations.
// This indicates a bug in LangExtract itself rather than user error.
type InternalError struct {
	*LangExtractError
}

func (e *InternalError) Error() string {
	return e.Message
}

// NewInternalError creates a new InternalError.
func NewInternalError(message string) *InternalError {
	return &InternalError{
		LangExtractError: NewLangExtractError(message),
	}
}

// ProviderError is a provider/backend specific error.
type ProviderError struct {
	*LangExtractError
	Provider string
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("provider %s: %s", e.Provider, e.Message)
}

// NewProviderError creates a new ProviderError.
func NewProviderError(provider, message string) *ProviderError {
	return &ProviderError{
		LangExtractError: NewLangExtractError(message),
		Provider:         provider,
	}
}

// SchemaError is a schema validation/serialization error.
type SchemaError struct {
	*LangExtractError
}

func (e *SchemaError) Error() string {
	return e.Message
}

// NewSchemaError creates a new SchemaError.
func NewSchemaError(message string) *SchemaError {
	return &SchemaError{
		LangExtractError: NewLangExtractError(message),
	}
}

// FormatError is the base error for format handling errors.
type FormatError struct {
	*LangExtractError
}

func (e *FormatError) Error() string {
	return e.Message
}

// NewFormatError creates a new FormatError.
func NewFormatError(message string) *FormatError {
	return &FormatError{
		LangExtractError: NewLangExtractError(message),
	}
}

// FormatParseError is raised when format parsing fails.
// This consolidates all parsing errors including:
// - Missing fence markers when required
// - Multiple fenced blocks
// - JSON/YAML decode errors
// - Missing wrapper keys
// - Invalid structure
type FormatParseError struct {
	*FormatError
}

func (e *FormatParseError) Error() string {
	return e.Message
}

// NewFormatParseError creates a new FormatParseError.
func NewFormatParseError(message string) *FormatParseError {
	return &FormatParseError{
		FormatError: NewFormatError(message),
	}
}

// Error type checking helpers
var (
	_ error = (*LangExtractError)(nil)
	_ error = (*InferenceError)(nil)
	_ error = (*InferenceConfigError)(nil)
	_ error = (*InferenceRuntimeError)(nil)
	_ error = (*InferenceOutputError)(nil)
	_ error = (*InvalidDocumentError)(nil)
	_ error = (*InternalError)(nil)
	_ error = (*ProviderError)(nil)
	_ error = (*SchemaError)(nil)
	_ error = (*FormatError)(nil)
	_ error = (*FormatParseError)(nil)
)

// IsLangExtractError checks if an error is a LangExtract error type.
func IsLangExtractError(err error) bool {
	var langExtractErr *LangExtractError
	return errors.As(err, &langExtractErr)
}

// IsInferenceError checks if an error is an inference error type.
func IsInferenceError(err error) bool {
	var inferenceErr *InferenceError
	return errors.As(err, &inferenceErr)
}

// IsInferenceConfigError checks if an error is a configuration error type.
func IsInferenceConfigError(err error) bool {
	var configErr *InferenceConfigError
	return errors.As(err, &configErr)
}

// IsInferenceRuntimeError checks if an error is a runtime error type.
func IsInferenceRuntimeError(err error) bool {
	var runtimeErr *InferenceRuntimeError
	return errors.As(err, &runtimeErr)
}

// IsInferenceOutputError checks if an error is an output error type.
func IsInferenceOutputError(err error) bool {
	var outputErr *InferenceOutputError
	return errors.As(err, &outputErr)
}

// IsInvalidDocumentError checks if an error is an invalid document error type.
func IsInvalidDocumentError(err error) bool {
	var docErr *InvalidDocumentError
	return errors.As(err, &docErr)
}

// IsInternalError checks if an error is an internal error type.
func IsInternalError(err error) bool {
	var internalErr *InternalError
	return errors.As(err, &internalErr)
}

// IsProviderError checks if an error is a provider error type.
func IsProviderError(err error) bool {
	var providerErr *ProviderError
	return errors.As(err, &providerErr)
}

// IsSchemaError checks if an error is a schema error type.
func IsSchemaError(err error) bool {
	var schemaErr *SchemaError
	return errors.As(err, &schemaErr)
}

// IsFormatError checks if an error is a format error type.
func IsFormatError(err error) bool {
	var formatErr *FormatError
	return errors.As(err, &formatErr)
}

// IsFormatParseError checks if an error is a format parse error type.
func IsFormatParseError(err error) bool {
	var parseErr *FormatParseError
	return errors.As(err, &parseErr)
}
