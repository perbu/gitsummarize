package summarizer

import "context"

type Summarizer interface {
	Summarize(ctx context.Context, commitMessages, diffs string) (string, error)
}
