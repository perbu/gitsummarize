package summarizer

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/ollama/ollama/api"
	"google.golang.org/api/option"
)

// Config holds the configuration for the summarizer.
type Config struct {
	UseOllama    bool
	OllamaModel  string
	GeminiAPIKey string
}

// New creates a new Summarizer based on the configuration.
func New(config Config) (Summarizer, error) {
	if config.UseOllama {
		return NewOllamaSummarizer(config.OllamaModel)
	}
	return NewGeminiSummarizer(config.GeminiAPIKey)
}

// GeminiSummarizer implements the Summarizer interface for Gemini API
type GeminiSummarizer struct {
	model *genai.GenerativeModel
}

// NewGeminiSummarizer creates a new GeminiSummarizer
func NewGeminiSummarizer(apiKey string) (*GeminiSummarizer, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Gemini API key is required")
	}
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	model := client.GenerativeModel("gemini-1.5-flash")
	return &GeminiSummarizer{model: model}, nil
}

// Summarize generates a summary for a given text using the Gemini API
func (s *GeminiSummarizer) Summarize(ctx context.Context, commitMessages, diffs string) (string, error) {
	prompt := fmt.Sprintf("Summarize the following git commits and their diffs in a single sentence, from the perspective of a tech lead reporting to management:\n\nCommits:\n%s\n\nDiffs:\n%s", commitMessages, diffs)

	slog.Debug("Gemini API request", "promptLength", len(prompt))

	startTime := time.Now()
	resp, err := s.model.GenerateContent(ctx, genai.Text(prompt))
	responseTime := time.Since(startTime)

	if err != nil {
		return "", err
	}

	promptTokens := 0
	candidatesTokens := 0
	if resp.UsageMetadata != nil {
		promptTokens = int(resp.UsageMetadata.PromptTokenCount)
		candidatesTokens = int(resp.UsageMetadata.CandidatesTokenCount)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		slog.Debug("Gemini API response", "status", "no summary generated", "responseTime", responseTime.String(), "promptTokens", promptTokens, "candidatesTokens", candidatesTokens)
		return "", fmt.Errorf("no summary generated")
	}

	summary := string(resp.Candidates[0].Content.Parts[0].(genai.Text))
	slog.Debug("Gemini API response", "status", "success", "summary", summary, "responseTime", responseTime.String(), "promptTokens", promptTokens, "candidatesTokens", candidatesTokens)

	return summary, nil
}

// OllamaSummarizer implements the Summarizer interface for Ollama API
type OllamaSummarizer struct {
	model  string
	client *api.Client
}

// NewOllamaSummarizer creates a new OllamaSummarizer
func NewOllamaSummarizer(model string) (*OllamaSummarizer, error) {
	// Get the Ollama host from the environment variable or use the default
	ollamaHost := os.Getenv("OLLAMA_HOST")
	if ollamaHost == "" {
		ollamaHost = "http://localhost:11434"
	}

	// Create a custom HTTP client with a timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create a new Ollama client with the custom HTTP client
	u, err := url.Parse(ollamaHost)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ollama host: %w", err)
	}
	client := api.NewClient(u, httpClient)

	return &OllamaSummarizer{model: model, client: client}, nil
}

// Summarize generates a summary for a given text using the Ollama API
func (s *OllamaSummarizer) Summarize(ctx context.Context, commitMessages, _ string) (string, error) {
	prompt := fmt.Sprintf("Summarize the following git commits in a single sentence, from the perspective of a tech lead reporting to management:\n\nCommits:\n%s", commitMessages)

	slog.Debug("Ollama API request", "promptLength", len(prompt), "model", s.model)

	startTime := time.Now()
	var fullResponse strings.Builder

	err := s.client.Generate(ctx, &api.GenerateRequest{
		Model:  s.model,
		Prompt: prompt,
	}, func(resp api.GenerateResponse) error {
		fullResponse.WriteString(resp.Response)
		return nil
	})
	responseTime := time.Since(startTime)

	if err != nil {
		return "", err
	}

	// Clean up the summary by removing the <think> tags and extra whitespace
	summary := fullResponse.String()
	re := regexp.MustCompile(`(?s)<think>.*</think>`)
	summary = re.ReplaceAllString(summary, "")
	summary = strings.TrimSpace(summary)

	if summary == "" {
		slog.Debug("Ollama API response", "status", "no summary generated", "responseTime", responseTime.String())
		return "", fmt.Errorf("no summary generated")
	}

	slog.Debug("Ollama API response", "status", "success", "summary", summary, "responseTime", responseTime.String())

	return summary, nil
}
