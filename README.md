# gitsummerize

`gitsummerize` is a command-line tool written in Go that summarizes git commits from a specified repository. It can aggregate commits by day and generate a concise summary for each day using either the Google Gemini API or a local Ollama instance. The output is a CSV file containing daily summaries, including commit counts, lines changed, and an AI-generated summary.

## Features

*   **Git Commit Aggregation:** Groups commits by day.
*   **AI-Powered Summarization:** Utilizes Google Gemini or Ollama to generate daily summaries of commit messages and diffs.
*   **Effort Calculation:** Provides a simple "effort" metric based on commit count and lines changed.
*   **CSV Output:** Generates a CSV report for easy analysis.
*   **Filtering:** Filter commits by author and date range.

## Installation

To build `gitsummerize` from source, you need Go installed (version 1.18 or higher).

```bash
git clone https://github.com/perbu/gitsummerize.git
cd gitsummerize
go build -o gitsummerize .
```

This will create an executable named `gitsummerize` in the current directory.

## Usage

```bash
./gitsummerize [flags]
```

### Flags

*   `--gemini-api-key string`: Your Google Gemini API key. Can also be set via the `GEMINI_API_KEY` environment variable.
*   `--use-ollama`: Use Ollama for summarization instead of Gemini.
*   `--ollama-model string`: The Ollama model to use (default: `qwen3:14b`). Only applicable when `--use-ollama` is set.
*   `--repo string`: Path to the git repository (default: `.` - current directory).
*   `--start-date string`: Optional start date in `YYYY-MM-DD` format to filter commits.
*   `--end-date string`: Optional end date in `YYYY-MM-DD` format to filter commits.
*   `--author string`: Optional author email to filter commits.

### Examples

**Summarize commits in the current repository using Gemini:**

```bash
./gitsummerize --gemini-api-key YOUR_GEMINI_API_KEY
```

**Summarize commits in a specific repository using Ollama:**

```bash
./gitsummerize --repo /path/to/your/repo --use-ollama --ollama-model llama3
```

**Summarize commits by a specific author within a date range:**

```bash
./gitsummerize --gemini-api-key YOUR_GEMINI_API_KEY --repo /path/to/your/repo --author "john.doe@example.com" --start-date "2024-01-01" --end-date "2024-01-31"
```

## Output

The tool outputs a CSV report to standard output (stdout) with the following columns:

*   `date`: The date of the daily summary (`YYYY-MM-DD`).
*   `effort in terms of days of work`: A calculated effort metric.
*   `no of commits`: Number of commits on that day.
*   `commit SHAs`: Space-separated list of commit SHAs.
*   `lines added`: Total lines added on that day.
*   `lines deleted`: Total lines deleted on that day.
*   `summary`: The AI-generated summary of the day's commits.

You can redirect the output to a file:

```bash
./gitsummerize --gemini-api-key YOUR_GEMINI_API_KEY > daily_report.csv
```

## Configuration

### Gemini API Key

You can provide your Gemini API key either via the `--gemini-api-key` flag or by setting the `GEMINI_API_KEY` environment variable.

### Ollama

If you choose to use Ollama, ensure you have Ollama installed and running, and that the specified model is available.
