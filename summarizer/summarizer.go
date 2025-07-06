package summarizer

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Summarize generates a summary for a given text using the Gemini API
func Summarize(apiKey, commitMessages, diffs string) (string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.5-flash")
	prompt := fmt.Sprintf("Summarize the following git commits and their diffs in a single sentence, from the perspective of a tech lead reporting to management:\n\nCommits:\n%s\n\nDiffs:\n%s", commitMessages, diffs)

	slog.Debug("Gemini API request",
		"prompt length", len(prompt),
	)

	startTime := time.Now()
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
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
		slog.Debug("Gemini API response",
			"status", "no summary generated",
			"responseTime", responseTime.String(),
			"promptTokens", promptTokens,
			"candidatesTokens", candidatesTokens,
		)
		return "", fmt.Errorf("no summary generated")
	}

	summary := string(resp.Candidates[0].Content.Parts[0].(genai.Text))
	slog.Debug("Gemini API response",
		"status", "success",
		"summary length", len(summary),
		"responseTime", responseTime.String(),
		"promptTokens", promptTokens,
		"candidatesTokens", candidatesTokens,
	)

	return summary, nil
}
