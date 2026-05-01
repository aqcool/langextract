package core

import (
	"errors"
	"testing"
)

func TestLangExtractError(t *testing.T) {
	err := NewLangExtractError("test error")
	if err.Error() != "test error" {
		t.Errorf("expected 'test error', got '%s'", err.Error())
	}
	if !IsLangExtractError(err) {
		t.Error("expected IsLangExtractError to return true")
	}
}

func TestInferenceConfigError(t *testing.T) {
	err := NewInferenceConfigError("missing api key")
	if !IsInferenceConfigError(err) {
		t.Error("expected IsInferenceConfigError to return true")
	}
	if !IsInferenceError(err) {
		t.Error("config error should also be an inference error")
	}
	if !IsLangExtractError(err) {
		t.Error("config error should also be a langextract error")
	}
}

func TestInferenceRuntimeError(t *testing.T) {
	orig := errors.New("connection refused")
	err := NewInferenceRuntimeError("API call failed", orig, "openai")
	if !IsInferenceRuntimeError(err) {
		t.Error("expected IsInferenceRuntimeError to return true")
	}
	if err.Unwrap() != orig {
		t.Error("Unwrap should return original error")
	}
}

func TestInferenceOutputError(t *testing.T) {
	err := NewInferenceOutputError("no output")
	if !IsInferenceOutputError(err) {
		t.Error("expected IsInferenceOutputError to return true")
	}
}

func TestInvalidDocumentError(t *testing.T) {
	err := NewInvalidDocumentError("empty text")
	if !IsInvalidDocumentError(err) {
		t.Error("expected IsInvalidDocumentError to return true")
	}
}

func TestFormatParseError(t *testing.T) {
	err := NewFormatParseError("invalid json")
	if !IsFormatParseError(err) {
		t.Error("expected IsFormatParseError to return true")
	}
	if !IsFormatError(err) {
		t.Error("parse error should also be a format error")
	}
}

func TestProviderError(t *testing.T) {
	err := NewProviderError("openai", "rate limited")
	if !IsProviderError(err) {
		t.Error("expected IsProviderError to return true")
	}
	if err.Provider != "openai" {
		t.Errorf("expected provider 'openai', got '%s'", err.Provider)
	}
}
