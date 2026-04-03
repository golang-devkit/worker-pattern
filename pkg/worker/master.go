package worker

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"syscall"
	"time"
)

// Option configures master server behavior
type Option struct {
	Args      []string
	NumWorker int
}

func (o Option) WithArgs(args ...string) Option {
	o.Args = args
	return o
}

func (o Option) WithNumWorker(n int) Option {
	o.NumWorker = n
	return o
}

func (o Option) Validate() error {
	if o.NumWorker <= 0 {
		return fmt.Errorf("NumWorker must be greater than 0")
	}
	return nil
}

func DefaultOption() Option {
	return Option{
		Args:      []string{},
		NumWorker: 5, // default
	}
}

func ValidateOptions(opts ...Option) error {
	for i, opt := range opts {
		if err := opt.Validate(); err != nil {
			return fmt.Errorf("Option %d validation failed: %v", i, err)
		}
	}
	return nil
}

// RunMasterServer starts the master server on the specified port with the given command and options.
// It creates a shared socket listener and spawns worker processes to handle incoming connections.
// The master process also handles graceful shutdown and notifies systemd of its status.
//
// This is the main entry point for the master server.
// It validates inputs, sets up the listener, spawns workers, and manages lifecycle events.
//
// Note: This function is designed to be simple and user-friendly, with sensible defaults and robust error handling.
// For more advanced use cases, consider using RunMasterWith() for greater control over the setup.
//
// Example usage:
//
//	RunMasterServer(8080, "myworker", DefaultOption().WithNumWorker(10).WithArgs("--verbose"))
//
// This will start a master server on port 8080, spawning 10 worker processes that run "myworker --verbose".
func RunMasterServer(port int, command string, opts ...Option) {

	// Validate port number
	if port <= 0 || port > 65535 {
		log.Fatalf("Invalid port number: %d", port)
	}

	// Validate command
	if command == "" {
		log.Fatal("Command cannot be empty")
	}

	// Set default options if none provided
	if opts == nil {
		opts = []Option{DefaultOption()}
	}

	//
	// Create shared socket listener (this is the key trick!)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Listening on port %d", port)

	// Get the file descriptor of the socket
	file, err := listener.(*net.TCPListener).File()
	if err != nil {
		log.Fatal(err)
	}

	RunMasterWith([]*os.File{file}, command, opts...)
}

// RunMasterWith starts the master server with the provided extra files (e.g., socket) and command,
// applying the given options.
// This function provides more flexibility than RunMasterServer(), allowing you to specify custom options
// and manage the setup in a more granular way.
//
// Example usage:
//
//	RunMasterWith([]*os.File{socketFile}, "myworker", DefaultOption().WithNumWorker(10).WithArgs("--verbose"))
//
// This will start a master server using the provided socket file, spawning 10 worker processes that run "myworker --verbose".
func RunMasterWith(extraFiles []*os.File, command string, opts ...Option) {

	// Validate extraFiles
	if len(extraFiles) == 0 {
		log.Fatal("At least one extra file (e.g., socket) must be provided")
	}

	// Validate command
	if command == "" {
		log.Fatal("Command cannot be empty")
	}

	// Set default options if none provided
	if opts == nil {
		opts = []Option{DefaultOption()}
	}

	// Override defaults with options
	//
	// Option 1: Only consider the first option struct (simpler)
	//
	// n := 5 // default
	// args := []string{}
	//
	// if len(opts) > 0 {
	// 	if opts[0].NumWorker > 0 {
	// 		n = opts[0].NumWorker
	// 	}
	// 	if opts[0].Args != nil {
	// 		args = opts[0].Args
	// 	}
	// }
	//

	// Option 2: Loop through all options and apply overrides (more flexible)
	//
	intArr := make([]int, len(opts))
	args := []string{}
	//
	// Loop through options and apply overrides
	for _, opt := range opts {
		// Override defaults if options provided
		if opt.NumWorker > 0 {
			intArr = append(intArr, opt.NumWorker)
		} else {
			intArr = append(intArr, 0) // n must be > 0 to override
		}
		// Append any additional arguments
		if len(opt.Args) > 0 {
			args = append(args, opt.Args...)
		}
	}

	//
	// Determine final worker count (use max of provided options or default)
	n := slices.Max(intArr)

	// Spawn worker processes
	sp := SpawnWorker(n, extraFiles, command, args...)

	// Wait for workers to start
	time.Sleep(500 * time.Millisecond)

	// Notify systemd: READY
	if err := sdNotify("READY=1"); err != nil {
		log.Printf("Warning: failed to notify systemd: %v", err)
	} else {
		log.Println("Notified systemd: READY")
	}
	// Write PID file for systemd
	pidFile := fmt.Sprintf("/run/%s.pid", filepath.Base(command))
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		log.Printf("Warning: Could not write PID file: %v", err)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Master received shutdown signal, stopping workers...")

	// Notify systemd: STOPPING
	if err := sdNotify("STOPPING=1"); err != nil {
		log.Printf("Warning: failed to notify systemd: %v", err)
	} else {
		log.Println("Notified systemd: STOPPING")
	}

	// Explicitly stop workers (not deferred)
	sp.Stop()

	// Remove PID file
	if err := os.Remove(pidFile); err != nil {
		log.Printf("Warning: Could not remove PID file: %v", err)
	}
	log.Println("Master process exiting")
}
