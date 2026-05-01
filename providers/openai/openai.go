// Package openai provides an OpenAI-compatible LLM provider.
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aqcool/langextract/core"
	"github.com/aqcool/langextract/providers"
)

const (
	// DefaultBaseURL is the default OpenAI API base URL.
	DefaultBaseURL = "https://api.openai.com/v1"
	
	// DefaultModelID is the default OpenAI model ID.
	DefaultModelID = "gpt-4o-mini"
	
	// DefaultMaxWorkers is the default number of parallel API calls.
	DefaultMaxWorkers = 10
	
	// DefaultTimeout is the default HTTP request timeout.
	DefaultTimeout = 120 * time.Second
	
	// MaxRetries is the maximum number of retries for API calls.
	MaxRetries = 3
	
	// RetryBaseDelay is the base delay for exponential backoff.
	RetryBaseDelay = 1 * time.Second
)

// OpenAIModel implements the BaseLanguageModel interface using OpenAI's API.
type OpenAIModel struct {
	client        *http.Client
	apiKey        string
	baseURL       string
	modelID       string
	organization  string
	formatType    core.FormatType
	temperature   *float64
	maxWorkers    int
	schema        providers.BaseSchema
	fenceOutput   *bool
	extraKwargs   map[string]interface{}
	mu            sync.Mutex
}

// Option is a functional option for OpenAIModel configuration.
type Option func(*OpenAIModel)

// WithAPIKey sets the API key.
func WithAPIKey(apiKey string) Option {
	return func(m *OpenAIModel) {
		m.apiKey = apiKey
	}
}

// WithBaseURL sets the base URL.
func WithBaseURL(baseURL string) Option {
	return func(m *OpenAIModel) {
		m.baseURL = baseURL
	}
}

// WithOrganization sets the organization ID.
func WithOrganization(org string) Option {
	return func(m *OpenAIModel) {
		m.organization = org
	}
}

// WithFormatType sets the output format type.
func WithFormatType(formatType core.FormatType) Option {
	return func(m *OpenAIModel) {
		m.formatType = formatType
	}
}

// WithTemperature sets the sampling temperature.
func WithTemperature(temp float64) Option {
	return func(m *OpenAIModel) {
		m.temperature = &temp
	}
}

// WithMaxWorkers sets the maximum number of parallel API calls.
func WithMaxWorkers(maxWorkers int) Option {
	return func(m *OpenAIModel) {
		m.maxWorkers = maxWorkers
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(m *OpenAIModel) {
		m.client = client
	}
}

// WithExtraKwargs sets extra keyword arguments.
func WithExtraKwargs(kwargs map[string]interface{}) Option {
	return func(m *OpenAIModel) {
		m.extraKwargs = kwargs
	}
}

// NewOpenAIModel creates a new OpenAIModel instance.
func NewOpenAIModel(modelID string, opts ...Option) (*OpenAIModel, error) {
	m := &OpenAIModel{
		modelID:     modelID,
		baseURL:     DefaultBaseURL,
		formatType:  core.FormatTypeJSON,
		maxWorkers:  DefaultMaxWorkers,
		extraKwargs: make(map[string]interface{}),
	}

	for _, opt := range opts {
		opt(m)
	}

	// Resolve API key from environment if not provided
	if m.apiKey == "" {
		m.apiKey = os.Getenv("OPENAI_API_KEY")
		if m.apiKey == "" {
			m.apiKey = os.Getenv("LANGEXTRACT_API_KEY")
		}
	}

	if m.apiKey == "" {
		return nil, core.NewInferenceConfigError("API key not provided. Set OPENAI_API_KEY or LANGEXTRACT_API_KEY environment variable, or pass WithAPIKey option.")
	}

	// Set default HTTP client if not provided
	if m.client == nil {
		m.client = &http.Client{
			Timeout: DefaultTimeout,
		}
	}

	return m, nil
}

// chatRequest represents an OpenAI chat completion request.
type chatRequest struct {
	Model          string                 `json:"model"`
	Messages       []chatMessage          `json:"messages"`
	N              int                    `json:"n,omitempty"`
	Temperature    *float64               `json:"temperature,omitempty"`
	MaxTokens      *int                   `json:"max_tokens,omitempty"`
	TopP           *float64               `json:"top_p,omitempty"`
	ResponseFormat *responseFormat        `json:"response_format,omitempty"`
	Stop           []string               `json:"stop,omitempty"`
	Seed           *int                   `json:"seed,omitempty"`
	FrequencyPenalty *float64             `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64             `json:"presence_penalty,omitempty"`
}

// chatMessage represents a chat message.
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// responseFormat represents the response format configuration.
type responseFormat struct {
	Type string `json:"type"`
}

// chatResponse represents an OpenAI chat completion response.
type chatResponse struct {
	ID      string       `json:"id"`
	Choices []chatChoice `json:"choices"`
	Usage   chatUsage    `json:"usage"`
}

// chatChoice represents a choice in the chat response.
type chatChoice struct {
	Index   int         `json:"index"`
	Message chatMessage `json:"message"`
}

// chatUsage represents token usage information.
type chatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// chatError represents an OpenAI API error response.
type chatError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// Call invokes the LLM with the given prompts and returns scored outputs.
func (m *OpenAIModel) Call(ctx context.Context, prompts []string, opts ...providers.CallOption) ([]core.ScoredOutput, error) {
	if len(prompts) == 0 {
		return nil, core.NewInferenceOutputError("no prompts provided")
	}

	// Apply call options
	callConfig := &providers.CallConfig{}
	for _, opt := range opts {
		opt(callConfig)
	}

	// Process prompts in parallel if multiple
	if len(prompts) > 1 && m.maxWorkers > 1 {
		return m.callParallel(ctx, prompts, callConfig)
	}

	// Process sequentially
	results := make([]core.ScoredOutput, len(prompts))
	for i, prompt := range prompts {
		output, err := m.processSinglePrompt(ctx, prompt, callConfig)
		if err != nil {
			return nil, err
		}
		results[i] = *output
	}

	return results, nil
}

// callParallel processes prompts in parallel using goroutines.
func (m *OpenAIModel) callParallel(ctx context.Context, prompts []string, callConfig *providers.CallConfig) ([]core.ScoredOutput, error) {
	numWorkers := m.maxWorkers
	if numWorkers > len(prompts) {
		numWorkers = len(prompts)
	}

	results := make([]core.ScoredOutput, len(prompts))
	errs := make([]error, len(prompts))
	
	var wg sync.WaitGroup
	sem := make(chan struct{}, numWorkers)

	for i, prompt := range prompts {
		wg.Add(1)
		sem <- struct{}{} // acquire semaphore
		
		go func(idx int, p string) {
			defer wg.Done()
			defer func() { <-sem }() // release semaphore
			
			output, err := m.processSinglePrompt(ctx, p, callConfig)
			if err != nil {
				errs[idx] = err
				return
			}
			results[idx] = *output
		}(i, prompt)
	}

	wg.Wait()

	// Check for errors
	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

// processSinglePrompt processes a single prompt and returns a ScoredOutput.
func (m *OpenAIModel) processSinglePrompt(ctx context.Context, prompt string, callConfig *providers.CallConfig) (*core.ScoredOutput, error) {
	var lastErr error
	
	for attempt := 0; attempt < MaxRetries; attempt++ {
		if attempt > 0 {
			delay := RetryBaseDelay * time.Duration(1<<uint(attempt-1))
			slog.Debug("Retrying OpenAI API call", "attempt", attempt+1, "delay", delay)
			
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}
		
		output, err := m.doAPICall(ctx, prompt, callConfig)
		if err != nil {
			lastErr = err
			// Retry on rate limit or server errors
			if isRetryableError(err) {
				continue
			}
			return nil, err
		}
		
		return output, nil
	}
	
	return nil, core.NewInferenceRuntimeError(
		fmt.Sprintf("OpenAI API call failed after %d retries: %v", MaxRetries, lastErr),
		lastErr,
		"openai",
	)
}

// doAPICall makes a single API call to OpenAI.
func (m *OpenAIModel) doAPICall(ctx context.Context, prompt string, callConfig *providers.CallConfig) (*core.ScoredOutput, error) {
	// Build system message based on format type
	var systemMessage string
	if m.formatType == core.FormatTypeJSON {
		systemMessage = "You are a helpful assistant that responds in JSON format."
	} else if m.formatType == core.FormatTypeYAML {
		systemMessage = "You are a helpful assistant that responds in YAML format."
	}

	// Build messages
	messages := []chatMessage{}
	if systemMessage != "" {
		messages = append(messages, chatMessage{Role: "system", Content: systemMessage})
	}
	messages = append(messages, chatMessage{Role: "user", Content: prompt})

	// Build request
	req := &chatRequest{
		Model:    m.modelID,
		Messages: messages,
		N:        1,
	}

	// Apply temperature
	temp := m.temperature
	if callConfig.Temperature != nil {
		temp = callConfig.Temperature
	}
	if temp != nil {
		req.Temperature = temp
	}

	// Apply max tokens
	if callConfig.MaxTokens != nil {
		req.MaxTokens = callConfig.MaxTokens
	}

	// Apply top_p
	if callConfig.TopP != nil {
		req.TopP = callConfig.TopP
	}

	// Apply stop sequences
	if len(callConfig.StopSequences) > 0 {
		req.Stop = callConfig.StopSequences
	}

	// Apply JSON mode (can be overridden via extra kwargs "response_format_type")
	if m.formatType == core.FormatTypeJSON {
		rfType := "json_object"
		if m.extraKwargs != nil {
			if v, ok := m.extraKwargs["response_format_type"].(string); ok && v != "" {
				rfType = v
			}
		}
		req.ResponseFormat = &responseFormat{Type: rfType}
	}

	// Apply extra kwargs
	if m.extraKwargs != nil {
		if v, ok := m.extraKwargs["max_output_tokens"]; ok {
			if val, ok := v.(int); ok {
				req.MaxTokens = &val
			}
		}
		if v, ok := m.extraKwargs["top_p"]; ok {
			if val, ok := v.(float64); ok {
				req.TopP = &val
			}
		}
		if v, ok := m.extraKwargs["frequency_penalty"]; ok {
			if val, ok := v.(float64); ok {
				req.FrequencyPenalty = &val
			}
		}
		if v, ok := m.extraKwargs["presence_penalty"]; ok {
			if val, ok := v.(float64); ok {
				req.PresencePenalty = &val
			}
		}
		if v, ok := m.extraKwargs["seed"]; ok {
			if val, ok := v.(int); ok {
				req.Seed = &val
			}
		}
	}

	// Marshal request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, core.NewInferenceRuntimeError(
			fmt.Sprintf("failed to marshal request: %v", err),
			err,
			"openai",
		)
	}

	// Create HTTP request
	url := strings.TrimRight(m.baseURL, "/") + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, core.NewInferenceRuntimeError(
			fmt.Sprintf("failed to create request: %v", err),
			err,
			"openai",
		)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+m.apiKey)
	if m.organization != "" {
		httpReq.Header.Set("OpenAI-Organization", m.organization)
	}

	// Send request
	resp, err := m.client.Do(httpReq)
	if err != nil {
		return nil, core.NewInferenceRuntimeError(
			fmt.Sprintf("OpenAI API request failed: %v", err),
			err,
			"openai",
		)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, core.NewInferenceRuntimeError(
			fmt.Sprintf("failed to read response: %v", err),
			err,
			"openai",
		)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var chatErr chatError
		if jsonErr := json.Unmarshal(respBody, &chatErr); jsonErr == nil && chatErr.Error.Message != "" {
			return nil, core.NewInferenceRuntimeError(
				fmt.Sprintf("OpenAI API error (status %d): %s", resp.StatusCode, chatErr.Error.Message),
				nil,
				"openai",
			)
		}
		return nil, core.NewInferenceRuntimeError(
			fmt.Sprintf("OpenAI API error (status %d): %s", resp.StatusCode, string(respBody)),
			nil,
			"openai",
		)
	}

	// Parse response
	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, core.NewInferenceRuntimeError(
			fmt.Sprintf("failed to parse response: %v", err),
			err,
			"openai",
		)
	}

	if len(chatResp.Choices) == 0 {
		return nil, core.NewInferenceOutputError("no choices in response")
	}

	outputText := chatResp.Choices[0].Message.Content
	score := 1.0

	return &core.ScoredOutput{
		Score:  &score,
		Output: &outputText,
	}, nil
}

// ModelID returns the model identifier.
func (m *OpenAIModel) ModelID() string {
	return m.modelID
}

// Provider returns the provider name.
func (m *OpenAIModel) Provider() string {
	return "openai"
}

// ApplySchema applies a schema instance to this provider.
func (m *OpenAIModel) ApplySchema(schema providers.BaseSchema) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.schema = schema
	return nil
}

// Schema returns the current schema instance if one is configured.
func (m *OpenAIModel) Schema() providers.BaseSchema {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.schema
}

// RequiresFenceOutput returns whether this model requires fence output for parsing.
func (m *OpenAIModel) RequiresFenceOutput() bool {
	if m.fenceOutput != nil {
		return *m.fenceOutput
	}
	// OpenAI JSON mode returns raw JSON without fences
	if m.formatType == core.FormatTypeJSON {
		return false
	}
	// For YAML or other formats, check schema
	if m.schema != nil && m.schema.RequiresRawOutput() {
		return false
	}
	return true
}

// SetFenceOutput sets explicit fence output preference.
func (m *OpenAIModel) SetFenceOutput(fenceOutput *bool) {
	m.fenceOutput = fenceOutput
}

// isRetryableError checks if an error is retryable (rate limit, server error).
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	// Check for InferenceRuntimeError with status code
	if runtimeErr, ok := err.(*core.InferenceRuntimeError); ok {
		msg := runtimeErr.Message
		// Rate limit errors
		if strings.Contains(msg, "status 429") {
			return true
		}
		// Server errors
		if strings.Contains(msg, "status 500") || strings.Contains(msg, "status 502") || strings.Contains(msg, "status 503") {
			return true
		}
	}
	
	return false
}
