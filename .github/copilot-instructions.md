# Copilot Instructions for worker-pattern

## Build, Test, and Run Commands

### Build the binary
```bash
make build
# Output: ./bin/app
```

### Run with default settings
```bash
make run
# Defaults to NumCPU workers, listens on port 8080
# Override: make run numWorkers=4 port=3000
```

### Run tests with race detector
```bash
make test
# Runs all Go tests across the entire module with -race flag
```

### Build for Linux deployment
```bash
make build-linux
# Output: ./bin/app_linux_amd64
```

### Clean artifacts
```bash
make clean
```

## High-Level Architecture

This is a **lab** researching the master-worker pattern in Go, deployed under systemd control. The goal is to enable multiple worker processes to share a single socket listener for efficient load distribution and graceful restarts.

### Research Structure

- **Experiments**: `cmd/v*/main.go` — Testbed implementations (line numbers and structure may change)
- **Results**: `pkg/worker/` — Stable, reusable package API extracted from successful experiments

### Core Concept

The socket-sharing trick works like this:
1. Master creates a TCP listener and extracts its file descriptor
2. Master passes the FD to worker child processes via `os.ExtraFiles` 
3. Workers reconstruct the listener from the inherited FD (typically FD 3)
4. All workers concurrently call `Accept()` on the shared socket
5. Kernel fairly distributes connections to available workers

### Key Components

**Master Process** (experimental in `cmd/v1/main.go`, stable API in `pkg/worker/master.go`)
- Creates TCP listener and extracts file descriptor
- Spawns N worker child processes, passing the FD via `ExtraFiles`
- Manages lifecycle: startup, graceful shutdown with timeout
- Notifies systemd of readiness and stopping states
- Writes/removes PID file for systemd tracking

**Worker Process** (experimental in `cmd/v1/main.go`, stable API in `pkg/worker/worker.go`)
- Receives inherited file descriptor and reconstructs the listener
- Each worker pinned to 1 CPU core via `runtime.GOMAXPROCS(1)`
- Serves HTTP requests from the shared socket
- Handles graceful shutdown with optional custom shutdown functions

**Supporting Modules** (all in `pkg/worker/`):
- `spawn.go` — Helper to spawn worker processes with environment and FD inheritance
- `notify.go` — systemd notification (READY, STOPPING signals)

### Execution Flow

1. Master starts → creates listener on TCP port
2. Master spawns N workers (default: NumCPU), passing socket FD via `ExtraFiles`
3. Each worker receives FD, reconstructs listener, serves HTTP independently
4. Kernel handles fair connection distribution across all workers
5. On graceful shutdown: Master receives SIGTERM → signals all workers → waits up to 30s → force kills if needed

## Key Conventions and Patterns

### File Descriptor Inheritance
- Master passes socket via `ExtraFiles` → appears as FD 3 in worker process
- Standard FDs: 0=stdin, 1=stdout, 2=stderr; ExtraFiles[0] → FD 3, ExtraFiles[1] → FD 4, etc.
- This pattern avoids port bind conflicts—only master binds; workers accept on inherited FD
- **See also**: `pkg/worker/worker.go` for how workers reconstruct the listener from FD 3

### Environment Variables
- `WORKER_ID`: Set by master on each worker process; used for logging and identification
- `NOTIFY_SOCKET`: Automatically set by systemd; enables systemd notifications

### Graceful Shutdown Pattern
- Signal handlers (SIGTERM/SIGINT) in both master and worker
- Master waits for all workers to exit, with configurable timeout (typically 30s) before force kill
- Workers execute custom shutdown functions if registered via `RegisterWorker()`
- See `pkg/worker/master.go` and `pkg/worker/worker.go` for current implementation

### CPU Affinity
- Each worker calls `runtime.GOMAXPROCS(1)` to pin to single core
- Reduces context switching and cache misses for CPU-bound operations

### systemd Integration
- Master notifies systemd readiness (READY=1) to prevent dependent services starting too early
- Graceful shutdown notification (STOPPING=1) sent before shutdown sequence
- PID file written to `/run/{binary_name}.pid` for systemd tracking
- See `pkg/worker/notify.go` for implementation

### Public API (pkg/worker/)
- `RunMasterServer(port, command, args)` — Simplest way to start master with child process
- `RunMasterWith(files, command, args)` — Advanced: pass custom file descriptors
- `RegisterWorker(executeFunc)` — Custom worker implementation with shutdown callback

## Code Navigation Tips

**Stable API Reference** (use these for production-ready examples):
- `pkg/worker/master.go` — `RunMasterServer()` and `RunMasterWith()` functions
- `pkg/worker/worker.go` — `RegisterWorker()` function for custom worker implementations
- `pkg/worker/examples/main.go` — Complete example of using the package API

**Experimental Implementations** (these are research testbeds in `cmd/v*/main.go`; structure and line numbers may change):
- `cmd/v1/main.go` — Full implementation combining master and worker in a single binary
- Look for: `runMaster()` function (socket creation, worker spawning, lifecycle management)
- Look for: `runWorker()` function (listener reconstruction, HTTP serving, graceful shutdown)
- Look for: `sdNotify()` function (systemd integration)

## GitNexus Code Intelligence

This project uses GitNexus for code analysis. Before editing any symbol, run:
```bash
gitnexus_impact({target: "functionName", direction: "upstream"})
```
to see the blast radius. See CLAUDE.md for full GitNexus guidance.

## MCP Server Configuration

For enhanced Copilot sessions, configure these MCP servers (see `.github/mcp-servers.json`):

- **bash**: Direct shell execution for running `make build`, `make test`, and other CLI commands
- **http**: Test HTTP endpoints locally (useful for testing `/` and `/heavy` routes on `localhost:8080`)

Use these to quickly verify builds, run tests, and validate HTTP endpoint responses without leaving the session.
