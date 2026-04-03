package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

var (
	numWorkers int
	port       int
)

func main() {
	// Kiểm tra xem đây là master hay worker process
	if os.Getenv("WORKER_ID") == "" {
		// Đọc cờ dòng lệnh
		flag.IntVar(&numWorkers, "worker", 0, "Number of worker processes")
		flag.IntVar(&port, "port", 8080, "Port to listen on")
		flag.Parse()
		// Mặc định sử dụng số CPU làm số worker
		if numWorkers <= 0 {
			numWorkers = runtime.NumCPU()
		}
		// MASTER PROCESS
		runMaster()
	} else {
		// WORKER PROCESS
		runWorker()
	}
}

func runMaster() {
	log.Println("Master process starting...")

	// Tạo shared socket listener (đây là trick quan trọng!)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}

	// Lấy file descriptor của socket
	file, err := listener.(*net.TCPListener).File()
	if err != nil {
		log.Fatal(err)
	}

	// Spawn worker processes
	workers := make([]*exec.Cmd, numWorkers)
	for i := 0; i < numWorkers; i++ {
		workers[i] = exec.Command(os.Args[0])
		workers[i].Env = append(os.Environ(),
			fmt.Sprintf("WORKER_ID=%d", i+1),
		)
		workers[i].Stdout = os.Stdout
		workers[i].Stderr = os.Stderr
		workers[i].ExtraFiles = []*os.File{file}

		if err := workers[i].Start(); err != nil {
			log.Printf("Failed to start worker %d: %v", i, err)
		} else {
			log.Printf("Worker %d started with PID %d", i, workers[i].Process.Pid)
		}
	}

	// Đợi workers khởi động
	time.Sleep(500 * time.Millisecond)

	// Thông báo cho systemd: READY
	if err := sdNotify("READY=1"); err != nil {
		log.Printf("Warning: failed to notify systemd: %v", err)
	} else {
		log.Println("Notified systemd: READY")
	}

	// Ghi PID file cho systemd
	pidFile := "/run/worker_pattern.pid"
	if err := os.WriteFile(pidFile, fmt.Appendf(nil, "%d", os.Getpid()), 0644); err != nil {
		log.Printf("Warning: Could not write PID file: %v", err)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Master received shutdown signal, stopping workers...")

	// Thông báo systemd đang stopping
	sdNotify("STOPPING=1")

	// Stop all workers gracefully
	for i, worker := range workers {
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
		for i, worker := range workers {
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
		for _, worker := range workers {
			if worker != nil && worker.Process != nil {
				worker.Process.Kill()
			}
		}
	}

	os.Remove(pidFile)
	log.Println("Master process exiting")
}

func runWorker() {
	workerID := os.Getenv("WORKER_ID")
	log.Printf("Worker %s (PID %d) starting...", workerID, os.Getpid())

	// Sử dụng 1 CPU core per worker
	runtime.GOMAXPROCS(1)

	// Lấy shared listener từ master
	// ExtraFiles[0] từ master tự động được map sang FD=3 trong worker process
	// (FD 0=stdin, 1=stdout, 2=stderr, 3=ExtraFiles[0], 4=ExtraFiles[1]...)
	listenerFile := os.NewFile(3, "listener")
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

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
		<-sigChan

		log.Printf("Worker %s shutting down...", workerID)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	// Start serving (tất cả workers cùng accept trên 1 socket)
	log.Printf("Worker %s are ready to connect via a shared listener from the master", workerID)
	if err := server.Serve(listener); err != http.ErrServerClosed {
		log.Printf("Worker %s error: %v", workerID, err)
	}

	log.Printf("Worker %s exited", workerID)
}

// sdNotify gửi thông báo đến systemd
func sdNotify(state string) error {
	socketPath := os.Getenv("NOTIFY_SOCKET")
	if socketPath == "" {
		// Không chạy dưới systemd, bỏ qua
		return nil
	}

	// Kết nối đến Unix socket của systemd
	conn, err := net.DialUnix("unixgram", nil, &net.UnixAddr{
		Name: socketPath,
		Net:  "unixgram",
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	// Gửi thông báo
	_, err = conn.Write([]byte(state))
	return err
}
