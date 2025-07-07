package summarizer

import (
	"bytes"
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

const (
	batchMaxSize    = 8192
	endOfDiffMarker = "---END-OF-DIFF---"
	batchPrompt     = `### Task
You are a code-review assistant. Summarize the **commit message** and **git diff** that follow. Each diff is separated by ---END-OF-DIFF---.

### Output style
- 3–5 bullet points, each ≤ 20 words  
- Only factual changes; no opinions, no future plans, no emojis  
- Use technical terms (API names, functions, files) as they appear  
- Use the past tense (“Added…”, “Fixed…”)  
- Omit author name, ticket numbers, and formatting details unless essential
- No markdown, no code blocks, no emojis, no links

### Input: 
`
	masterSummary = `
### Task
You are a release-notes assistant. Produce one concise **master summary** that unifies the commit summaries supplied.

### Output instructions
- 4 – 7 bullet points, each ≤ 25 words  
- Facts only; no opinions or next-step wording  
- Merge duplicates, group related changes, keep highest-level view  
- Keep technical terms (APIs, functions, files) exactly as written  
- Use past tense (“Added…”, “Refactored…”)
- No markdown, no code blocks, no emojis, no links

### Input:
`
)

// BatchedOllamaSummarizer implements the Summarizer interface for Ollama API
type BatchedOllamaSummarizer struct {
	model  string
	client *api.Client
}

// NewBatchedOllamaSummarizer creates a new BatchedOllamaSummarizer
func NewBatchedOllamaSummarizer(model string) (*BatchedOllamaSummarizer, error) {
	// Get the Ollama host from the environment variable or use the default
	ollamaHost := os.Getenv("OLLAMA_HOST")
	if ollamaHost == "" {
		ollamaHost = "http://localhost:11434"
	}

	// Create a custom HTTP client with a timeout
	httpClient := &http.Client{
		Timeout: 180 * time.Second,
	}

	// Create a new Ollama client with the custom HTTP client
	u, err := url.Parse(ollamaHost)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ollama host: %w", err)
	}
	client := api.NewClient(u, httpClient)

	return &BatchedOllamaSummarizer{model: model, client: client}, nil
}

// Summarize generates a summary for a given text using the Ollama API
func (s *BatchedOllamaSummarizer) Summarize(ctx context.Context, commits []git.Commit) (string, error) {
	var summaries []string
	var buffer bytes.Buffer

	for _, commit := range commits {
		if buffer.Len()+len(commit.Message) > batchMaxSize {
			prompt := batchPrompt + buffer.String()
			summary, err := s.generateSummary(ctx, prompt)
			if err != nil {
				return "", err
			}
			summaries = append(summaries, summary)
			buffer.Reset()
		}
		buffer.WriteString(commit.Message)
		buffer.WriteString("\n")
		buffer.WriteString(commit.Diff)
		buffer.WriteString("\n")
		buffer.WriteString(endOfDiffMarker)
		buffer.WriteString("\n")
	}

	if buffer.Len() > 0 {
		prompt := masterSummary + buffer.String()
		summary, err := s.generateSummary(ctx, prompt)
		if err != nil {
			return "", err
		}
		summaries = append(summaries, summary)
	}

	finalText := strings.Join(summaries, "\n")
	prompt := fmt.Sprintf("Summarize the following git commit summaries in a single sentence, from the perspective of a tech lead reporting to management: %s", finalText)
	return s.generateSummary(ctx, prompt)
}

func (s *BatchedOllamaSummarizer) generateSummary(ctx context.Context, prompt string) (string, error) {
	slog.Debug("Ollama API request", "promptLength", len(prompt), "model", s.model)

	startTime := time.Now()
	var fullResponse strings.Builder

	var promptTokens, responseTokens int
	elements := 0
	err := s.client.Generate(ctx, &api.GenerateRequest{
		Model:  s.model,
		Prompt: prompt,
	}, func(resp api.GenerateResponse) error {
		elements++
		fullResponse.WriteString(resp.Response)
		if resp.Done {
			promptTokens = resp.PromptEvalCount
			responseTokens = resp.EvalCount
			slog.Debug("Ollama API response", "responseTime", time.Since(startTime).String(), "promptTokens", promptTokens, "responseTokens", responseTokens, "elements", elements)
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
	re := regexp.MustCompile(`(?s)<think>.*?</think>`)
	summary = re.ReplaceAllString(summary, "")
	summary = strings.TrimSpace(summary)

	if summary == "" {
		slog.Debug("Ollama API response", "status", "no summary generated", "responseTime", responseTime.String(), "promptTokens", promptTokens, "responseTokens", responseTokens)
		return "", fmt.Errorf("no summary generated")
	}

	slog.Debug("Ollama API response", "status", "success", "summary", summary, "responseTime", responseTime.String(), "promptTokens", promptTokens, "responseTokens", responseTokens)

	return summary, nil
}
