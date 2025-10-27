// jobqueue_test.go
package jobqueue

import (
	"sync"
	"testing"
	"time"
)

func TestJobQueueBasicProcessing(t *testing.T) {
	jq := NewJobQueue(3)
	processed := make(map[int]bool)
	var processedMu sync.Mutex

	jq.Start(func(j Job) {
		time.Sleep(10 * time.Millisecond) // simulate work
		processedMu.Lock()
		processed[j.ID] = true
		processedMu.Unlock()
	})

	jobs := []Job{
		{ID: 1, Payload: "a"},
		{ID: 2, Payload: "b"},
		{ID: 3, Payload: "c"},
		{ID: 4, Payload: "d"},
	}
	for _, job := range jobs {
		jq.AddJob(job)
	}

	jq.WaitAll()

	processedMu.Lock()
	if len(processed) != 4 {
		t.Errorf("expected all jobs processed, got %d", len(processed))
	}
	for _, job := range jobs {
		if !processed[job.ID] {
			t.Errorf("job %d not processed", job.ID)
		}
	}
	processedMu.Unlock()
}

func TestJobQueueProgress(t *testing.T) {
	jq := NewJobQueue(2)
	jq.Start(func(j Job) { time.Sleep(20 * time.Millisecond) })
	jobs := []Job{
		{ID: 1, Payload: "a"},
		{ID: 2, Payload: "b"},
		{ID: 3, Payload: "c"},
	}
	for _, job := range jobs {
		jq.AddJob(job)
	}
	time.Sleep(10 * time.Millisecond)
	completed, total := jq.Progress()
	if completed < 0 || total != 3 {
		t.Errorf("unexpected progress: %d/%d", completed, total)
	}
	jq.WaitAll()
	completed, total = jq.Progress()
	if completed != 3 || total != 3 {
		t.Errorf("after wait, unexpected progress: %d/%d", completed, total)
	}
}

func TestJobQueueNoDuplicatesOrMisses(t *testing.T) {
	jq := NewJobQueue(5)
	var mu sync.Mutex
	seen := make(map[int]struct{})
	jq.Start(func(j Job) {
		mu.Lock()
		if _, exists := seen[j.ID]; exists {
			t.Errorf("duplicate job picked up: %d", j.ID)
		}
		seen[j.ID] = struct{}{}
		mu.Unlock()
		// Simulate variable work
		time.Sleep(time.Millisecond * time.Duration(5+j.ID%3))
	})
	for i := 0; i < 15; i++ {
		jq.AddJob(Job{ID: i, Payload: "p"})
	}
	jq.WaitAll()
	mu.Lock()
	if len(seen) != 15 {
		t.Errorf("expected 15 jobs, got %d", len(seen))
	}
	mu.Unlock()
}
