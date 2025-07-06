package main

import (
	"flag"
	"log/slog"
	"os"
	"strings"

	"gitsummerize/git"
	"gitsummerize/report"
	"gitsummerize/summarizer"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	geminiAPIKey := flag.String("gemini-api-key", "", "Google Gemini API key")
	repoPath := flag.String("repo", ".", "path to the git repository")
	startDate := flag.String("start-date", "", "optional start date in YYYY-MM-DD format")
	endDate := flag.String("end-date", "", "optional end date in YYYY-MM-DD format")
	author := flag.String("author", "", "optional author email to filter commits")
	flag.Parse()

	slog.Info("starting gitsummerize",
		"repo", *repoPath,
		"startDate", *startDate,
		"endDate", *endDate,
		"author", *author,
	)

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = *geminiAPIKey
	}

	if apiKey == "" {
		slog.Error("Gemini API key is required. Set it with the GEMINI_API_KEY environment variable or the --gemini-api-key flag.")
		os.Exit(1)
	}

	commits, err := git.GetCommits(*repoPath, *author, *startDate, *endDate)
	if err != nil {
		slog.Error("failed to get commits", "err", err)
		os.Exit(1)
	}

	dailySummaries := report.AggregateByDay(commits)

	for i := range dailySummaries {
		var commitMessages []string
		var diffs []string
		for _, commit := range dailySummaries[i].Commits {
			commitMessages = append(commitMessages, commit.Message)
			diffs = append(diffs, commit.Diff)
		}

		summary, err := summarizer.Summarize(apiKey, strings.Join(commitMessages, "\n"), strings.Join(diffs, "\n"))
		if err != nil {
			slog.Error("failed to generate summary", "err", err)
			continue
		}
		dailySummaries[i].Summary = summary
	}

	report.CalculateEffort(dailySummaries)

	if err := report.WriteCSV(dailySummaries, os.Stdout); err != nil {
		slog.Error("failed to write CSV", "err", err)
		os.Exit(1)
	}

	slog.Info("gitsummerize finished successfully")
}
