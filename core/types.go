// Package core provides core data types and interfaces for LangExtract.
package core

// FormatType represents the output format type for prompts.
type FormatType string

const (
	FormatTypeYAML FormatType = "yaml"
	FormatTypeJSON FormatType = "json"
)

// ConstraintType represents the type of constraint for model output decoding.
type ConstraintType string

const (
	ConstraintTypeNone ConstraintType = "none"
)

// Constraint represents a constraint for model output decoding.
type Constraint struct {
	ConstraintType ConstraintType `json:"constraint_type"`
}

// ScoredOutput represents scored output from language model inference.
type ScoredOutput struct {
	Score  *float64 `json:"score,omitempty"`
	Output *string  `json:"output,omitempty"`
}

// String returns a formatted string representation of the ScoredOutput.
func (s ScoredOutput) String() string {
	scoreStr := "-"
	if s.Score != nil {
		scoreStr = "%.2f"
	}

	if s.Output == nil {
		return "Score: " + scoreStr + "\nOutput: None"
	}

	return "Score: " + scoreStr + "\nOutput:\n  " + *s.Output
}
