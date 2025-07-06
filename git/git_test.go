package git

import (
	"os"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestGetCommits(t *testing.T) {
	// Create a temporary directory for the test repository
	t.Parallel()
	dir, err := os.MkdirTemp("", "test-repo")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Initialize a new git repository
	r, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new file and commit it
	_, err = os.Create(dir + "/test-file")
	if err != nil {
		t.Fatal(err)
	}

	w, err := r.Worktree()
	if err != nil {
		t.Fatal(err)
	}

	_, err = w.Add("test-file")
	if err != nil {
		t.Fatal(err)
	}

	_, err = w.Commit("initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Get the commits
	commits, err := GetCommits(dir, "", "", "")
	if err != nil {
		t.Fatal(err)
	}

	// Check that there is one commit
	if len(commits) != 1 {
		t.Errorf("expected 1 commit, got %d", len(commits))
	}

	// Check the commit author
	if commits[0].Author != "test@example.com" {
		t.Errorf("expected author 'test@example.com', got '%s'", commits[0].Author)
	}
}
