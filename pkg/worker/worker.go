package worker

import (
	"context"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func RegisterWorker(execute func(fdShared *os.File, workerID string,
	graceful func(shutdownFunc func(context.Context) error))) {

	// Lấy WORKER_ID từ biến môi trường
	workerID := os.Getenv("WORKER_ID")
	log.Printf("Worker %s (PID %d) starting...", workerID, os.Getpid())

	// Sử dụng 1 CPU core per worker
	runtime.GOMAXPROCS(1)

	// Lấy shared listener từ master
	listenerFile := os.NewFile(3, "listener") // FD 3 vì 0=stdin, 1=stdout, 2=stderr, 3=first ExtraFile, 4=second, ...

	// Đăng ký hàm thiết lập shutdown
	shutdown := func(ctx context.Context) error {
		log.Printf("Worker %s default shutdown (no-op)", workerID)
		return nil
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
		<-sigChan

		log.Printf("Worker %s shutting down...", workerID)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		shutdown(ctx)
	}()

	// Chạy server worker với listener và hàm thiết lập shutdown
	execute(listenerFile, workerID, func(fn func(context.Context) error) {
		// Đăng ký hàm shutdown do worker cung cấp
		shutdown = fn
	})

	log.Printf("Worker %s exited", workerID)
}
