package worker

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// SpawnWorker starts n worker processes with the given command and arguments,
// passing the extraFiles (e.g., shared socket) to each worker.
func SpawnWorker(n int, extraFiles []*os.File,
	name string, arg ...string) *Spawn {

	// Spawn worker processes
	workers := make([]*exec.Cmd, n)
	for i := range n {
		workers[i] = exec.Command(name, arg...)
		workers[i].Env = append(os.Environ(), fmt.Sprintf("WORKER_ID=%d", i+1))
		workers[i].Stdout = os.Stdout
		workers[i].Stderr = os.Stderr
		workers[i].ExtraFiles = extraFiles

		if err := workers[i].Start(); err != nil {
			log.Printf("Failed to start worker %d: %v", i, err)
		} else {
			log.Printf("Worker %d started with PID %d", i, workers[i].Process.Pid)
		}
	}

	return &Spawn{
		workers: workers,
	}
}

type Spawn struct {
	workers []*exec.Cmd
}

func (s *Spawn) Stop() {
	// Stop all workers gracefully
	for i, worker := range s.workers {
		if worker != nil && worker.Process != nil {
			log.Printf("Stopping worker %d (PID %d)", i, worker.Process.Pid)
			if err := worker.Process.Signal(syscall.SIGTERM); err != nil {
				log.Printf("Warning: failed to signal worker %d (PID %d): %v",
					i, worker.Process.Pid, err)
			}
		}
	}
	// Wait with timeout
	done := make(chan struct{})
	go func() {
		for i, worker := range s.workers {
			if worker != nil {
				worker.Wait()
				log.Printf("Worker %d stopped", i)
			}
		}
		close(done)
	}()
	select {
	case <-done:
		log.Println("All workers stopped gracefully")
	case <-time.After(30 * time.Second):
		log.Println("Timeout waiting for workers, forcing shutdown")
		for _, worker := range s.workers {
			if worker != nil && worker.Process != nil {
				worker.Process.Kill()
			}
		}
	}
}
