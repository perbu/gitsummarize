package git

import (
	"bytes"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Commit represents a single git commit
type Commit struct {
	Hash    string
	Author  string
	Date    time.Time
	Message string
	Added   int
	Deleted int
	Diff    string
}

// GetCommits opens a git repository and returns a slice of commits
func GetCommits(repoPath, author, startDate, endDate string) ([]Commit, error) {
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, err
	}

	cIter, err := r.Log(&git.LogOptions{All: true})
	if err != nil {
		return nil, err
	}

	var commits []Commit
	err = cIter.ForEach(func(c *object.Commit) error {
		// Author filter
		if author != "" && c.Author.Email != author {
			return nil
		}

		// Date filter
		if startDate != "" {
			start, err := time.Parse("2006-01-02", startDate)
			if err == nil && c.Author.When.Before(start) {
				return nil
			}
		}
		if endDate != "" {
			end, err := time.Parse("2006-01-02", endDate)
			if err == nil && c.Author.When.After(end) {
				return nil
			}
		}

		stats, err := c.Stats()
		if err != nil {
			return err
		}

		var added, deleted int
		for _, stat := range stats {
			added += stat.Addition
			deleted += stat.Deletion
		}

		parent, err := c.Parent(0)
		if err != nil {
			// This is likely the initial commit, which has no parent.
			// We can treat the diff as the full content of the commit.
			parent = nil
		}

		patch, err := c.Patch(parent)
		if err != nil {
			return err
		}

		var diff bytes.Buffer
		if err := patch.Encode(&diff); err != nil {
			return err
		}

		commits = append(commits, Commit{
			Hash:    c.Hash.String(),
			Author:  c.Author.Email,
			Date:    c.Author.When,
			Message: c.Message,
			Added:   added,
			Deleted: deleted,
			Diff:    diff.String(),
		})
		return nil
	})

	if err != nil {
		return nil, err
	}

	return commits, nil
}
