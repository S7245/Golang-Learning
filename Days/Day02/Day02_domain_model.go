package main

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

type JobID string
type WorkerID string

type JobStatus string

const (
	StatusPending   JobStatus = "pending"
	StatusQueued    JobStatus = "queued"
	StatusRunning   JobStatus = "running"
	StatusSucceeded JobStatus = "succeeded"
	StatusFailed    JobStatus = "failed"
	StatusCanceled  JobStatus = "canceled"
)

var (
	ErrEmptyJobID        = errors.New("job id cannot be empty")
	ErrEmptyJobTitle     = errors.New("job title cannot be empty")
	ErrEmptyWorkerID     = errors.New("worker id cannot be empty")
	ErrInvalidMaxRetry   = errors.New("max retry must be zero or greater")
	ErrInvalidStatus     = errors.New("invalid job status")
	ErrInvalidTransition = errors.New("invalid job status transition")
)

type JobSpec struct {
	ID         JobID
	Title      string
	QueueName  string
	MaxRetry   int
	CreatedAt  time.Time
	Attributes map[string]string
}

type Job struct {
	ID         JobID
	Title      string
	QueueName  string
	Status     JobStatus
	MaxRetry   int
	Attempts   int
	CreatedAt  time.Time
	StartedAt  time.Time
	FinishedAt time.Time
	WorkerID   WorkerID
	Attributes map[string]string
}

type Worker struct {
	ID        WorkerID
	Name      string
	CreatedAt time.Time
}

type Queue struct {
	Name string
	jobs []Job
}

func NewJob(spec JobSpec) (Job, error) {
	id := JobID(strings.TrimSpace(string(spec.ID)))
	if id == "" {
		return Job{}, ErrEmptyJobID
	}

	title := strings.TrimSpace(spec.Title)
	if title == "" {
		return Job{}, ErrEmptyJobTitle
	}

	if spec.MaxRetry < 0 {
		return Job{}, ErrInvalidMaxRetry
	}

	now := spec.CreatedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}

	queueName := strings.TrimSpace(spec.QueueName)
	if queueName == "" {
		queueName = "default"
	}

	return Job{
		ID:         id,
		Title:      title,
		QueueName:  queueName,
		Status:     StatusPending,
		MaxRetry:   spec.MaxRetry,
		CreatedAt:  now,
		Attributes: cloneMap(spec.Attributes),
	}, nil
}

func NewWorker(id WorkerID, name string, createdAt time.Time) (Worker, error) {
	id = WorkerID(strings.TrimSpace(string(id)))
	if id == "" {
		return Worker{}, ErrEmptyWorkerID
	}

	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	return Worker{
		ID:        id,
		Name:      strings.TrimSpace(name),
		CreatedAt: createdAt,
	}, nil
}

func (j Job) IsZero() bool {
	return j.ID == "" && j.Title == "" && j.Status == ""
}

func (j Job) CanRetry() bool {
	return j.Status == StatusFailed && j.Attempts <= j.MaxRetry
}

func (j Job) Age(now time.Time) time.Duration {
	if j.CreatedAt.IsZero() || now.Before(j.CreatedAt) {
		return 0
	}
	return now.Sub(j.CreatedAt)
}

func (j Job) Summary() string {
	worker := "unassigned"
	if j.WorkerID != "" {
		worker = string(j.WorkerID)
	}
	return fmt.Sprintf("%s [%s] %s queue=%s attempts=%d/%d worker=%s", j.ID, j.Status, j.Title, j.QueueName, j.Attempts, j.MaxRetry, worker)
}

func (j *Job) Enqueue() error {
	return j.transition(StatusQueued, time.Time{})
}

func (j *Job) Start(worker Worker, now time.Time) error {
	if worker.ID == "" {
		return ErrEmptyWorkerID
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if err := j.transition(StatusRunning, now); err != nil {
		return err
	}
	j.WorkerID = worker.ID
	j.Attempts++
	return nil
}

func (j *Job) Complete(now time.Time) error {
	return j.finish(StatusSucceeded, now)
}

func (j *Job) Fail(now time.Time) error {
	return j.finish(StatusFailed, now)
}

func (j *Job) Cancel(now time.Time) error {
	return j.finish(StatusCanceled, now)
}

func (j *Job) Retry(now time.Time) error {
	if !j.CanRetry() {
		return ErrInvalidTransition
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	j.FinishedAt = time.Time{}
	return j.transition(StatusQueued, now)
}

func (j *Job) finish(next JobStatus, now time.Time) error {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if err := j.transition(next, now); err != nil {
		return err
	}
	j.FinishedAt = now
	return nil
}

func (j *Job) transition(next JobStatus, at time.Time) error {
	if !next.Valid() {
		return ErrInvalidStatus
	}
	if !canTransition(j.Status, next) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, j.Status, next)
	}
	j.Status = next
	if next == StatusRunning && !at.IsZero() {
		j.StartedAt = at
	}
	return nil
}

func (s JobStatus) Valid() bool {
	switch s {
	case StatusPending, StatusQueued, StatusRunning, StatusSucceeded, StatusFailed, StatusCanceled:
		return true
	default:
		return false
	}
}

func canTransition(current, next JobStatus) bool {
	switch current {
	case StatusPending:
		return next == StatusQueued || next == StatusCanceled
	case StatusQueued:
		return next == StatusRunning || next == StatusCanceled
	case StatusRunning:
		return next == StatusSucceeded || next == StatusFailed || next == StatusCanceled
	case StatusFailed:
		return next == StatusQueued
	default:
		return false
	}
}

func (q *Queue) Enqueue(job Job) error {
	if job.IsZero() {
		return ErrEmptyJobID
	}
	if job.Status == StatusPending {
		if err := job.Enqueue(); err != nil {
			return err
		}
	}
	q.jobs = append(q.jobs, job)
	return nil
}

func (q Queue) Jobs() []Job {
	copied := make([]Job, len(q.jobs))
	copy(copied, q.jobs)
	return copied
}

func (q Queue) ByStatus(status JobStatus) []Job {
	var filtered []Job
	for _, job := range q.jobs {
		if job.Status == status {
			filtered = append(filtered, job)
		}
	}
	return filtered
}

func (q Queue) StatusCounts() map[JobStatus]int {
	counts := map[JobStatus]int{}
	for _, job := range q.jobs {
		counts[job.Status]++
	}
	return counts
}

func cloneMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	copied := make(map[string]string, len(values))
	for key, value := range values {
		copied[key] = value
	}
	return copied
}

func sampleQueue(now time.Time) (Queue, error) {
	queue := Queue{Name: "default"}
	jobs := []JobSpec{
		{
			ID:        "job-001",
			Title:     "import customer tasks",
			QueueName: "default",
			MaxRetry:  2,
			CreatedAt: now.Add(-20 * time.Minute),
			Attributes: map[string]string{
				"source": "csv",
				"owner":  "ops",
			},
		},
		{
			ID:        "job-002",
			Title:     "send workflow summary",
			QueueName: "notifications",
			MaxRetry:  1,
			CreatedAt: now.Add(-10 * time.Minute),
		},
	}

	for _, spec := range jobs {
		job, err := NewJob(spec)
		if err != nil {
			return Queue{}, err
		}
		if err := queue.Enqueue(job); err != nil {
			return Queue{}, err
		}
	}
	return queue, nil
}

func printJobs(jobs []Job, now time.Time) {
	if len(jobs) == 0 {
		fmt.Println("no jobs")
		return
	}
	for _, job := range jobs {
		fmt.Printf("%s age=%s\n", job.Summary(), job.Age(now).Round(time.Second))
	}
}

func printCounts(counts map[JobStatus]int) {
	statuses := make([]string, 0, len(counts))
	for status := range counts {
		statuses = append(statuses, string(status))
	}
	sort.Strings(statuses)
	for _, status := range statuses {
		fmt.Printf("%s=%d\n", status, counts[JobStatus(status)])
	}
}

func run(args []string, now time.Time) error {
	queue, err := sampleQueue(now)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		args = []string{"demo"}
	}

	switch args[0] {
	case "demo":
		worker, err := NewWorker("worker-001", "local runner", now)
		if err != nil {
			return err
		}
		jobs := queue.Jobs()
		if err := jobs[0].Start(worker, now.Add(30*time.Second)); err != nil {
			return err
		}
		if err := jobs[0].Complete(now.Add(2 * time.Minute)); err != nil {
			return err
		}
		printJobs(jobs, now.Add(3*time.Minute))
	case "list":
		printJobs(queue.Jobs(), now)
	case "counts":
		printCounts(queue.StatusCounts())
	case "invalid":
		job, err := NewJob(JobSpec{ID: "bad-001", Title: "invalid transition", CreatedAt: now})
		if err != nil {
			return err
		}
		return job.Complete(now)
	case "help":
		fmt.Println("commands: demo, list, counts, invalid")
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
	return nil
}

func main() {
	now := time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC)
	if err := run(os.Args[1:], now); err != nil {
		fmt.Fprintln(os.Stderr, "gotaskflow:", err)
		os.Exit(1)
	}
}
