package summarizer

import (
	"context"
	"fmt"

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

	model := client.GenerativeModel("gemini-1.5-flash")
	prompt := fmt.Sprintf("Summarize the following git commits and their diffs in a single sentence, from the perspective of a tech lead reporting to management:\n\nCommits:\n%s\n\nDiffs:\n%s", commitMessages, diffs)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no summary generated")
	}

	return string(resp.Candidates[0].Content.Parts[0].(genai.Text)), nil
}
