package summarizer

import (
	"context"
	"github.com/perbu/gitsummarize/git"
)

type Summarizer interface {
	Summarize(ctx context.Context, commits []git.Commit) (string, error)
}
