package queue

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type Uploader interface {
	Upload(ctx context.Context, filePath string) error
}

type Task struct {
	FilePath string
	Retries  int
}

type Queue struct {
	uploader   Uploader
	maxRetries int
	tasks      chan Task
	wg         sync.WaitGroup
	failedDir  string
}

func NewQueue(uploader Uploader, maxRetries int, monitorDir string) (*Queue, error) {
	failedDir := filepath.Join(monitorDir, ".failed_uploads")
	if err := os.MkdirAll(failedDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create failed dir: %w", err)
	}

	q := &Queue{
		uploader:   uploader,
		maxRetries: maxRetries,
		tasks:      make(chan Task, 100),
		failedDir:  failedDir,
	}

	q.wg.Add(1)
	go q.worker()

	return q, nil
}

func (q *Queue) Enqueue(filePath string) {
	q.tasks <- Task{
		FilePath: filePath,
		Retries:  0,
	}
}

func (q *Queue) worker() {
	defer q.wg.Done()

	for task := range q.tasks {
		if err := q.processTask(task); err != nil {
			log.Printf("Task failed: %v", err)
		}
	}
}

func (q *Queue) processTask(task Task) error {
	ctx := context.Background()
	err := q.uploader.Upload(ctx, task.FilePath)

	if err != nil {
		if task.Retries < q.maxRetries {
			task.Retries++
			log.Printf("Upload failed, retrying (%d/%d): %s",
				task.Retries, q.maxRetries, task.FilePath)
			q.tasks <- task
			return nil
		}

		return q.moveToFailed(task.FilePath)
	}

	log.Printf("Successfully uploaded: %s", task.FilePath)
	return nil
}

func (q *Queue) moveToFailed(filePath string) error {
	fileName := filepath.Base(filePath)
	destPath := filepath.Join(q.failedDir, fileName)

	if err := os.Rename(filePath, destPath); err != nil {
		return fmt.Errorf("failed to move to failed dir: %w", err)
	}

	log.Printf("Moved to failed directory: %s", destPath)
	return nil
}

func (q *Queue) Close() {
	close(q.tasks)
	q.wg.Wait()
}
