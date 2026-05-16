package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

const modulePath = "github.com/S7245/Golang-Learning"

type LayoutEntry struct {
	Path           string
	Boundary       string
	Responsibility string
}

type JobStatus string

const (
	JobQueued JobStatus = "queued"
	JobDone   JobStatus = "done"
)

type Job struct {
	ID        string
	Title     string
	Status    JobStatus
	CreatedAt time.Time
}

type Queue struct {
	jobs []Job
}

func (q *Queue) Add(title string, now time.Time) (Job, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Job{}, errors.New("job title cannot be empty")
	}

	job := Job{
		ID:        fmt.Sprintf("job-%03d", len(q.jobs)+1),
		Title:     title,
		Status:    JobQueued,
		CreatedAt: now,
	}
	q.jobs = append(q.jobs, job)
	return job, nil
}

func (q Queue) List() []Job {
	copied := make([]Job, len(q.jobs))
	copy(copied, q.jobs)
	return copied
}

func projectLayout() []LayoutEntry {
	return []LayoutEntry{
		{
			Path:           "go.mod",
			Boundary:       "module",
			Responsibility: "declares the stable module path and Go version for GoTaskFlow",
		},
		{
			Path:           "cmd/gotaskflow/main.go",
			Boundary:       "executable entry",
			Responsibility: "parses CLI input, assembles dependencies, and reports user-facing errors",
		},
		{
			Path:           "internal/taskflow",
			Boundary:       "private domain package",
			Responsibility: "owns jobs, queues, workers, and workflow rules that are not public API",
		},
		{
			Path:           "Days/Day01",
			Boundary:       "learning artifact",
			Responsibility: "keeps the daily tutorial and runnable sample isolated from production code",
		},
	}
}

func validateLayout(entries []LayoutEntry) error {
	seen := map[string]bool{}
	for _, entry := range entries {
		if strings.TrimSpace(entry.Path) == "" {
			return errors.New("layout entry has empty path")
		}
		if seen[entry.Path] {
			return fmt.Errorf("layout entry %q is duplicated", entry.Path)
		}
		seen[entry.Path] = true
	}
	return nil
}

func printLayout(entries []LayoutEntry) {
	fmt.Printf("GoTaskFlow module: %s\n\n", modulePath)
	for _, entry := range entries {
		fmt.Printf("%-28s %-20s %s\n", entry.Path, entry.Boundary, entry.Responsibility)
	}
}

func printJobs(jobs []Job) {
	if len(jobs) == 0 {
		fmt.Println("queue is empty")
		return
	}
	for _, job := range jobs {
		fmt.Printf("%s [%s] %s\n", job.ID, job.Status, job.Title)
	}
}

func run(args []string, now time.Time) error {
	layout := projectLayout()
	if err := validateLayout(layout); err != nil {
		return err
	}

	queue := Queue{}
	if len(args) == 0 || args[0] == "layout" {
		printLayout(layout)
		return nil
	}

	switch args[0] {
	case "add":
		if len(args) < 2 {
			return errors.New("usage: go run Days/Day01/Day01_module_layout.go add \"task title\"")
		}
		job, err := queue.Add(strings.Join(args[1:], " "), now)
		if err != nil {
			return err
		}
		fmt.Printf("added %s: %s\n", job.ID, job.Title)
	case "list":
		printJobs(queue.List())
	case "help":
		fmt.Println("commands: layout, add <title>, list")
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
	return nil
}

func main() {
	if err := run(os.Args[1:], time.Now()); err != nil {
		fmt.Fprintln(os.Stderr, "gotaskflow:", err)
		os.Exit(1)
	}
}
