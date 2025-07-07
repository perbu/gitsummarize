package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"gitsummerize/git"
	"gitsummerize/report"
	"gitsummerize/summarizer"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	geminiAPIKey := flag.String("gemini-api-key", "", "Google Gemini API key")
	useOllama := flag.Bool("use-ollama", false, "Use Ollama for summarization")
	useBatchedOllama := flag.Bool("use-batched-ollama", false, "Use batched Ollama for summarization")
	ollamaModel := flag.String("ollama-model", "qwen3:32b", "Ollama model to use")
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
		"useOllama", *useOllama,
		"useBatchedOllama", *useBatchedOllama,
		"ollamaModel", *ollamaModel,
	)

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = *geminiAPIKey
	}

	summarizerClient, err := summarizer.New(summarizer.Config{
		UseOllama:        *useOllama,
		UseBatchedOllama: *useBatchedOllama,
		OllamaModel:      *ollamaModel,
		GeminiAPIKey:     apiKey,
	})
	if err != nil {
		slog.Error("failed to create summarizer", "err", err)
		os.Exit(1)
	}

	commits, err := git.GetCommits(*repoPath, *author, *startDate, *endDate)
	if err != nil {
		slog.Error("failed to get commits", "err", err)
		os.Exit(1)
	}

	dailySummaries := report.AggregateByDay(commits)
	slog.Info("generated daily summaries", "count", len(dailySummaries))

	for i := range dailySummaries {
		slog.Debug("generating summary", "date", dailySummaries[i].Date.Format("2006-01-02"))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		summary, err := summarizerClient.Summarize(ctx, dailySummaries[i].Commits)
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
