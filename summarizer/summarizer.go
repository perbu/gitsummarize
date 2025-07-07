package summarizer

// Config holds the configuration for the summarizer.
type Config struct {
	UseOllama        bool
	UseBatchedOllama bool
	OllamaModel      string
	GeminiAPIKey     string
}

// New creates a new Summarizer based on the configuration.
func New(config Config) (Summarizer, error) {
	if config.UseBatchedOllama {
		return NewBatchedOllamaSummarizer(config.OllamaModel)
	}
	if config.UseOllama {
		return NewOllamaSummarizer(config.OllamaModel)
	}
	return NewGeminiSummarizer(config.GeminiAPIKey)
}
