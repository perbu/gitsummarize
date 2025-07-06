package report

import (
	"encoding/csv"
	"io"
	"strconv"
	"strings"
	"time"

	"gitsummerize/git"
)

// DailySummary represents the summary of a single day
type DailySummary struct {
	Date         time.Time
	Commits      []git.Commit
	CommitCount  int
	LinesAdded   int
	LinesDeleted int
	Summary      string
	Effort       float64
}

// AggregateByDay aggregates commits by day
func AggregateByDay(commits []git.Commit) []DailySummary {
	// Create a map to store daily summaries
	summaries := make(map[time.Time]DailySummary)

	// Iterate over the commits and add them to the corresponding daily summary
	for _, commit := range commits {
		// Get the date of the commit
		date := commit.Date.Truncate(24 * time.Hour)

		// Get the daily summary for the date
		summary := summaries[date]

		// Update the daily summary
		summary.Date = date
		summary.Commits = append(summary.Commits, commit)
		summary.CommitCount++
		summary.LinesAdded += commit.Added
		summary.LinesDeleted += commit.Deleted

		// Save the updated daily summary
		summaries[date] = summary
	}

	// Create a slice of daily summaries
	result := make([]DailySummary, 0, len(summaries))

	// Add the daily summaries to the slice
	for _, summary := range summaries {
		result = append(result, summary)
	}

	// Return the slice of daily summaries
	return result
}

// CalculateEffort calculates the effort for each daily summary
func CalculateEffort(summaries []DailySummary) {
	for i := range summaries {
		summaries[i].Effort = float64(summaries[i].CommitCount+((summaries[i].LinesAdded+summaries[i].LinesDeleted)/100)) / 10
	}
}

// WriteCSV writes the daily summaries to a CSV file
func WriteCSV(summaries []DailySummary, w io.Writer) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{"date", "effort in terms of days of work", "no of commits", "commit SHAs", "lines added", "lines deleted", "summary"}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, summary := range summaries {
		var commitSHAs []string
		for _, commit := range summary.Commits {
			commitSHAs = append(commitSHAs, commit.Hash)
		}

		record := []string{
			summary.Date.Format("2006-01-02"),
			strconv.FormatFloat(summary.Effort, 'f', 2, 64),
			strconv.Itoa(summary.CommitCount),
			strings.Join(commitSHAs, " "),
			strconv.Itoa(summary.LinesAdded),
			strconv.Itoa(summary.LinesDeleted),
			summary.Summary,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}
