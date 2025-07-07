package summarizer

import (
	"context"
	"gitsummerize/git"
)

type Summarizer interface {
	Summarize(ctx context.Context, commits []git.Commit) (string, error)
}
