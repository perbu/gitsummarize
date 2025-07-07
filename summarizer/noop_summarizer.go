package summarizer

import (
	"context"
	"gitsummerize/git"
)

// NoopSummarizer is a summarizer that does nothing.
type NoopSummarizer struct{}

// NewNoopSummarizer creates a new NoopSummarizer.
func NewNoopSummarizer() (*NoopSummarizer, error) {
	return &NoopSummarizer{}, nil
}

// Summarize returns an empty string.
func (s *NoopSummarizer) Summarize(_ context.Context, _ []git.Commit) (string, error) {
	return "", nil
}
