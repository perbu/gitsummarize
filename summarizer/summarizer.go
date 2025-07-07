package summarizer

import "strings"

// Config holds the configuration for the summarizer.
type Config struct {
	Summarize bool
	Batched   bool
	Model     string

	GeminiAPIKey string
}

// New creates a new Summarizer based on the configuration.
func New(config Config) (Summarizer, error) {
	if !config.Summarize {
		return NewNoopSummarizer()
	}
	if config.Batched {
		return NewBatchedOllamaSummarizer(config.Model)
	}
	if strings.HasPrefix(config.Model, "gemini") {
		return NewGeminiSummarizer(config.GeminiAPIKey)
	}
	return NewOllamaSummarizer(config.Model)
}
