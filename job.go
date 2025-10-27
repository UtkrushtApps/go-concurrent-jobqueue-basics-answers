// job.go
package jobqueue

// Job represents a unit of work to process.
type Job struct {
	ID      int
	Payload string
}
