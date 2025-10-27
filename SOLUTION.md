# Solution Steps

1. Define the Job struct in job.go to uniquely identify each job and carry a payload.

2. In jobqueue.go, implement the JobQueue struct with fields: the queue (slice), inProgress map, completed map, mutex, cond, and worker count.

3. In NewJobQueue, setup the maps, slice, and initialize a sync.Cond lock for coordination.

4. Implement AddJob with mutex protection and signal via cond when a new job is added.

5. Create fetchJobLocked for dequeuing jobs (under lock) and tracking as in-progress.

6. Implement the worker goroutine method: repeatedly wait for available jobs, take and unlock, process, then mark as completed and remove from inProgress (protected by lock).

7. Implement Start to launch N worker goroutines, each using the processing function provided.

8. Progress returns completed and total-count (add up completed, in-progress, and queued jobs); protect state with the mutex.

9. WaitAll spins/waits until all queued and in-progress jobs are processed.

10. In jobqueue_test.go, write table-driven and concurrency tests: add several jobs, process them concurrently, check for correct progress, ensure no duplicate/missing jobs via maps, and test WaitAll blocks till all are done.

