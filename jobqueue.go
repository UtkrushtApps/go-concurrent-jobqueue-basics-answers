// jobqueue.go
package jobqueue

import (
	"sync"
)

// JobQueue manages concurrent job processing.
type JobQueue struct {
	queue      []Job
	inProgress map[int]Job
	completed  map[int]Job
	mu         sync.Mutex
	cond       *sync.Cond
	workers    int
}

// NewJobQueue initializes a job queue with 'numWorkers'.
func NewJobQueue(numWorkers int) *JobQueue {
	jq := &JobQueue{
		queue:      make([]Job, 0),
		inProgress: make(map[int]Job),
		completed:  make(map[int]Job),
		workers:    numWorkers,
	}
	jq.cond = sync.NewCond(&jq.mu)
	return jq
}

// AddJob safely appends a job to the queue.
func (jq *JobQueue) AddJob(job Job) {
	jq.mu.Lock()
	defer jq.mu.Unlock()
	jq.queue = append(jq.queue, job)
	jq.cond.Signal()
}

// fetchJobLocked fetches the next job from the queue under lock.
func (jq *JobQueue) fetchJobLocked() (Job, bool) {
	if len(jq.queue) == 0 {
		return Job{}, false
	}
	job := jq.queue[0]
	jq.queue = jq.queue[1:]
	jq.inProgress[job.ID] = job
	return job, true
}

// worker is run as a goroutine and continually processes jobs.
func (jq *JobQueue) worker(process func(Job)) {
	for {
		jq.mu.Lock()
		for len(jq.queue) == 0 {
			jq.cond.Wait()
		}
		job, ok := jq.fetchJobLocked()
		jq.mu.Unlock()
		if ok {
			process(job)
			jq.mu.Lock()
			delete(jq.inProgress, job.ID)
			jq.completed[job.ID] = job
			jq.mu.Unlock()
		}
	}
}

// Start launches all workers which run the provided process function.
func (jq *JobQueue) Start(process func(Job)) {
	for i := 0; i < jq.workers; i++ {
		go jq.worker(process)
	}
}

// Progress returns (completed, total) jobs.
func (jq *JobQueue) Progress() (int, int) {
	jq.mu.Lock()
	defer jq.mu.Unlock()
	completed := len(jq.completed)
	total := completed + len(jq.inProgress) + len(jq.queue)
	return completed, total
}

// WaitAll blocks until all jobs are completed.
func (jq *JobQueue) WaitAll() {
	for {
		jq.mu.Lock()
		completed := len(jq.completed)
		total := completed + len(jq.inProgress) + len(jq.queue)
		jq.mu.Unlock()
		if completed == total {
			return
		}
	}
}
