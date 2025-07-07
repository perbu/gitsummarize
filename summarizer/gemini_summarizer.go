package summarizer

import (
	"context"
	"fmt"
	"github.com/perbu/gitsummarize/git"
	"log/slog"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

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
func (s *GeminiSummarizer) Summarize(ctx context.Context, commits []git.Commit) (string, error) {
	var commitMessages, diffs []string
	for _, commit := range commits {
		commitMessages = append(commitMessages, commit.Message)
		diffs = append(diffs, commit.Diff)
	}
	prompt := fmt.Sprintf("Summarize the following git commits and their diffs in a single sentence, from the perspective of a tech lead reporting to management:\n\nCommits:\n%s\n\nDiffs:\n%s", strings.Join(commitMessages, "\n---\n"), strings.Join(diffs, "\n---\n"))

	slog.Debug("Gemini API request", "promptLength", len(prompt))

	startTime := time.Now()
	resp, err := s.model.GenerateContent(ctx, genai.Text(prompt))
	responseTime := time.Since(startTime)

	if err != nil {
		slog.Error("Gemini API request failed", "err", err, "responseTime", responseTime.String())
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
