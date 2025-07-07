package summarizer

import (
	"context"
	"fmt"
	"gitsummerize/git"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/ollama/ollama/api"
)

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
func (s *OllamaSummarizer) Summarize(ctx context.Context, commits []git.Commit) (string, error) {
	var commitMessages []string
	for _, commit := range commits {
		commitMessages = append(commitMessages, commit.Message)
	}
	prompt := fmt.Sprintf("Summarize the following git commits in a single sentence, from the perspective of a tech lead reporting to management:\n\nCommits:\n%s", strings.Join(commitMessages, "\n---\n"))

	slog.Debug("Ollama API request", "promptLength", len(prompt), "model", s.model)

	startTime := time.Now()
	var fullResponse strings.Builder

	var promptTokens, responseTokens int
	err := s.client.Generate(ctx, &api.GenerateRequest{
		Model:  s.model,
		Prompt: prompt,
	}, func(resp api.GenerateResponse) error {
		fullResponse.WriteString(resp.Response)
		if resp.Done {
			promptTokens = resp.PromptEvalCount
			responseTokens = resp.EvalCount
		}
		return nil
	})
	responseTime := time.Since(startTime)

	if err != nil {
		slog.Error("Ollama API request failed", "err", err, "responseTime", responseTime.String())
		return "", err
	}

	// Clean up the summary by removing the <think> tags and extra whitespace
	summary := fullResponse.String()
	re := regexp.MustCompile(`(?s)<think>.*</think>`)
	summary = re.ReplaceAllString(summary, "")
	summary = strings.TrimSpace(summary)

	if summary == "" {
		slog.Debug("Ollama API response", "status", "no summary generated", "responseTime", responseTime.String(), "promptTokens", promptTokens, "responseTokens", responseTokens)
		return "", fmt.Errorf("no summary generated")
	}

	slog.Debug("Ollama API response", "status", "success", "summary", summary, "responseTime", responseTime.String(), "promptTokens", promptTokens, "responseTokens", responseTokens)

	return summary, nil
}
