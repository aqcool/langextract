package core

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// Constants for extraction data keys.
const (
	ExtractionsKey = "extractions"
	AttributeSuffix = "_attributes"
)

// AlignmentStatus represents the alignment status of an extraction.
type AlignmentStatus string

const (
	AlignmentStatusMatchExact   AlignmentStatus = "match_exact"
	AlignmentStatusMatchGreater AlignmentStatus = "match_greater"
	AlignmentStatusMatchLesser  AlignmentStatus = "match_lesser"
	AlignmentStatusMatchFuzzy   AlignmentStatus = "match_fuzzy"
)

// CharInterval represents a character interval in the original text.
type CharInterval struct {
	StartPos *int `json:"start_pos,omitempty"`
	EndPos   *int `json:"end_pos,omitempty"`
}

// TokenInterval represents an interval over tokens in tokenized text.
type TokenInterval struct {
	StartIndex int `json:"start_index"`
	EndIndex   int `json:"end_index"`
}

// Extraction represents an extraction extracted from text.
// This class encapsulates an extraction's characteristics and its position
// within the source text. It can represent a diverse range of information for
// NLP information extraction tasks.
type Extraction struct {
	ExtractionClass  string                 `json:"extraction_class"`
	ExtractionText   string                 `json:"extraction_text"`
	CharInterval     *CharInterval          `json:"char_interval,omitempty"`
	AlignmentStatus  AlignmentStatus        `json:"alignment_status,omitempty"`
	ExtractionIndex  *int                   `json:"extraction_index,omitempty"`
	GroupIndex       *int                   `json:"group_index,omitempty"`
	Description      *string                `json:"description,omitempty"`
	Attributes       map[string]interface{} `json:"attributes,omitempty"`
	TokenInterval    *TokenInterval         `json:"token_interval,omitempty"`
}

// Document represents a document for annotation.
type Document struct {
	Text             string  `json:"text"`
	DocumentID       string  `json:"document_id"`
	AdditionalContext *string `json:"additional_context,omitempty"`
}

// NewDocument creates a new Document with auto-generated ID if not provided.
func NewDocument(text string, opts ...DocumentOption) *Document {
	doc := &Document{
		Text: text,
	}
	
	for _, opt := range opts {
		opt(doc)
	}
	
	// Auto-generate document ID if not set
	if doc.DocumentID == "" {
		doc.DocumentID = fmt.Sprintf("doc_%s", uuid.New().String()[:8])
	}
	
	return doc
}

// DocumentOption is a functional option for Document configuration.
type DocumentOption func(*Document)

// WithDocumentID sets the document ID.
func WithDocumentID(id string) DocumentOption {
	return func(d *Document) {
		d.DocumentID = id
	}
}

// WithAdditionalContext sets the additional context.
func WithAdditionalContext(context string) DocumentOption {
	return func(d *Document) {
		d.AdditionalContext = &context
	}
}

// AnnotatedDocument represents a document with extractions.
type AnnotatedDocument struct {
	DocumentID   string        `json:"document_id"`
	Text         *string       `json:"text,omitempty"`
	Extractions  []*Extraction `json:"extractions,omitempty"`
}

// NewAnnotatedDocument creates a new AnnotatedDocument with auto-generated ID if not provided.
func NewAnnotatedDocument(opts ...AnnotatedDocumentOption) *AnnotatedDocument {
	doc := &AnnotatedDocument{}
	
	for _, opt := range opts {
		opt(doc)
	}
	
	// Auto-generate document ID if not set
	if doc.DocumentID == "" {
		doc.DocumentID = fmt.Sprintf("doc_%s", uuid.New().String()[:8])
	}
	
	return doc
}

// AnnotatedDocumentOption is a functional option for AnnotatedDocument configuration.
type AnnotatedDocumentOption func(*AnnotatedDocument)

// WithAnnotatedDocumentID sets the document ID.
func WithAnnotatedDocumentID(id string) AnnotatedDocumentOption {
	return func(d *AnnotatedDocument) {
		d.DocumentID = id
	}
}

// WithAnnotatedDocumentText sets the text.
func WithAnnotatedDocumentText(text string) AnnotatedDocumentOption {
	return func(d *AnnotatedDocument) {
		d.Text = &text
	}
}

// WithExtractions sets the extractions.
func WithExtractions(extractions []*Extraction) AnnotatedDocumentOption {
	return func(d *AnnotatedDocument) {
		d.Extractions = extractions
	}
}

// ExampleData represents a single training/example data instance for structured prompting.
type ExampleData struct {
	Text        string        `json:"text"`
	Extractions []*Extraction `json:"extractions,omitempty"`
}

// NewExampleData creates a new ExampleData instance.
func NewExampleData(text string, extractions ...*Extraction) *ExampleData {
	return &ExampleData{
		Text:        text,
		Extractions: extractions,
	}
}

// ToJSON converts the extraction to JSON bytes.
func (e *Extraction) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON creates an Extraction from JSON bytes.
func (e *Extraction) FromJSON(data []byte) error {
	return json.Unmarshal(data, e)
}

// ToJSON converts the document to JSON bytes.
func (d *Document) ToJSON() ([]byte, error) {
	return json.Marshal(d)
}

// FromJSON creates a Document from JSON bytes.
func (d *Document) FromJSON(data []byte) error {
	return json.Unmarshal(data, d)
}

// ToJSON converts the annotated document to JSON bytes.
func (ad *AnnotatedDocument) ToJSON() ([]byte, error) {
	return json.Marshal(ad)
}

// FromJSON creates an AnnotatedDocument from JSON bytes.
func (ad *AnnotatedDocument) FromJSON(data []byte) error {
	return json.Unmarshal(data, ad)
}

// ToJSON converts the example data to JSON bytes.
func (ed *ExampleData) ToJSON() ([]byte, error) {
	return json.Marshal(ed)
}

// FromJSON creates an ExampleData from JSON bytes.
func (ed *ExampleData) FromJSON(data []byte) error {
	return json.Unmarshal(data, ed)
}

// Clone creates a deep copy of the Extraction.
func (e *Extraction) Clone() *Extraction {
	if e == nil {
		return nil
	}
	
	clone := &Extraction{
		ExtractionClass: e.ExtractionClass,
		ExtractionText:  e.ExtractionText,
		AlignmentStatus: e.AlignmentStatus,
	}
	
	if e.CharInterval != nil {
		ci := &CharInterval{}
		if e.CharInterval.StartPos != nil {
			v := *e.CharInterval.StartPos
			ci.StartPos = &v
		}
		if e.CharInterval.EndPos != nil {
			v := *e.CharInterval.EndPos
			ci.EndPos = &v
		}
		clone.CharInterval = ci
	}
	
	if e.ExtractionIndex != nil {
		idx := *e.ExtractionIndex
		clone.ExtractionIndex = &idx
	}
	
	if e.GroupIndex != nil {
		idx := *e.GroupIndex
		clone.GroupIndex = &idx
	}
	
	if e.Description != nil {
		desc := *e.Description
		clone.Description = &desc
	}
	
	if e.Attributes != nil {
		clone.Attributes = make(map[string]interface{})
		for k, v := range e.Attributes {
			clone.Attributes[k] = v
		}
	}
	
	if e.TokenInterval != nil {
		clone.TokenInterval = &TokenInterval{
			StartIndex: e.TokenInterval.StartIndex,
			EndIndex:   e.TokenInterval.EndIndex,
		}
	}
	
	return clone
}

// Clone creates a deep copy of the Document.
func (d *Document) Clone() *Document {
	if d == nil {
		return nil
	}
	
	clone := &Document{
		Text:       d.Text,
		DocumentID: d.DocumentID,
	}
	
	if d.AdditionalContext != nil {
		ctx := *d.AdditionalContext
		clone.AdditionalContext = &ctx
	}
	
	return clone
}

// Clone creates a deep copy of the AnnotatedDocument.
func (ad *AnnotatedDocument) Clone() *AnnotatedDocument {
	if ad == nil {
		return nil
	}
	
	clone := &AnnotatedDocument{
		DocumentID: ad.DocumentID,
	}
	
	if ad.Text != nil {
		text := *ad.Text
		clone.Text = &text
	}
	
	if ad.Extractions != nil {
		clone.Extractions = make([]*Extraction, len(ad.Extractions))
		for i, e := range ad.Extractions {
			clone.Extractions[i] = e.Clone()
		}
	}
	
	return clone
}
