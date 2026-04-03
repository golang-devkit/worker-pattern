package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/golang-devkit/worker-pattern/pkg/worker"
)

func main() {
	// Kiểm tra xem đây là master hay worker process
	if os.Getenv("WORKER_ID") == "" {
		// Example 1: Default behavior (backward compatible, 5 workers)
		worker.RunMasterServer(8080, os.Args[0])

		// Example 2: With custom worker count (uncomment to use)
		// worker.RunMasterServer(8080, os.Args[0], worker.Option{NumWorker: 4})

		// Example 3: With both custom worker count and arguments
		// worker.RunMasterServer(8080, os.Args[0], worker.Option{
		// 	NumWorker: 8,
		// 	Args:      []string{"-verbose"},
		// })
	} else {
		worker.RegisterWorker(runWorker)
	}
}

func runWorker(listenerFile *os.File, workerID string,
	graceful func(func(context.Context) error)) {

	listener, err := net.FileListener(listenerFile)
	if err != nil {
		log.Fatalf("Worker %s: failed to get listener: %v", workerID, err)
	}

	// Tạo HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Handled by worker %s (PID %d)\n", workerID, os.Getpid())
	})

	mux.HandleFunc("/heavy", func(w http.ResponseWriter, r *http.Request) {
		// Simulate CPU-intensive work
		start := time.Now()
		sum := 0
		for i := range 100000000 {
			sum += i
		}
		fmt.Fprintf(w, "Worker %s processed heavy request in %v\n",
			workerID, time.Since(start))
	})

	server := &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	// Listen for shutdown signal
	graceful(server.Shutdown)

	// Start serving (tất cả workers cùng accept trên 1 socket)
	log.Printf("Worker %s are ready to connect via a shared listener from the master", workerID)
	if err := server.Serve(listener); err != http.ErrServerClosed {
		log.Printf("Worker %s error: %v", workerID, err)
	}
}
