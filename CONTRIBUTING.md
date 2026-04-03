# Contributing to worker-pattern Lab

<!-- markdownlint-disable MD013 MD025 MD033 MD060 -->

Thank you for your interest in contributing to this research lab! This guide explains how to effectively participate in developing the worker-pattern implementation.

## What is This Lab?

This is a **research project** investigating the master-worker pattern for monolithic applications deployed under systemd control. The goal is to develop a reusable, production-ready Go package (`pkg/worker/`) by testing implementations in isolated experiments (`cmd/v*/main.go`).

### Experiment vs. Results Structure

- **Experiments** (`cmd/v*/main.go`): Testbeds for trying new approaches. Code here is unstable and expected to change frequently.
- **Results** (`pkg/worker/`): Stable, production-ready package extracted from successful experiments. Changes here must maintain API stability.

Contributions may target either area, but treat them differently:

- Experiments can be aggressive; refactoring, restructuring, and line changes are expected
- Package code requires backward compatibility and API review

## Getting Started

### Prerequisites

- Go 1.25.7+ (see `go.mod`)
- Unix/Linux system (for systemd integration testing)
- Basic understanding of process management and socket programming

### Local Setup

```bash
# Clone the repository
git clone <repo-url>
cd lab_WorkerPattern

# Install dependencies (vendored via Makefile)
make fetch-module

# Build the application
make build

# Run tests
make test
```

## Development Workflow

### Before You Start

1. **Read the README.md** to understand the lab's goals and structure
2. **Check existing issues** to avoid duplicate work
3. **Review pkg/worker/ API** in `pkg/worker/master.go`, `pkg/worker/worker.go`, and `pkg/worker/examples/main.go` if your change affects the public API

### Making Changes

#### For Experiments (cmd/v*/main.go)

- Create new versions as needed (e.g., `cmd/v2/main.go`)
- Document your experiment goal in comments
- Don't worry about breaking existing experiments—they're testbeds
- Update README.md if adding a new version with distinct research goals

Example:

```go
// cmd/v2/main.go - Testing improved worker scheduling with affinity pinning
```

#### For Package Code (pkg/worker/)

- **Changes must maintain backward compatibility** for public functions
- Use GitNexus to check impact before modifying exported symbols:

  ```bash
  gitnexus_impact({target: "FunctionName", direction: "upstream"})
  ```

- Update examples in `pkg/worker/examples/main.go` if the API changes
- Add inline comments for non-obvious design decisions

#### For All Code

- Follow Go conventions:
  - Use `gofmt` for formatting (automatic via `make build`)
  - Use meaningful variable and function names
  - Avoid unexported helper functions unless truly internal
  - Keep comments minimal; prefer clear code over verbose comments

- Test your changes:

  ```bash
  make test        # Run all tests with race detector
  make clean       # Clean build artifacts before committing
  ```

## Public API (pkg/worker/)

### Configuration with Option Struct

The package provides an `Option` struct for configurable server behavior:

```go
type Option struct {
    Args      []string  // Additional arguments to pass to workers
    NumWorker int32     // Number of worker processes (default: 5)
}
```

### API Functions

- `RunMasterServer(port int, command string, opts ...Option)` — Start master with optional configuration
- `RunMasterWith(extraFiles []*os.File, command string, opts ...Option)` — Advanced: custom file descriptors
- `RegisterWorker(executeFunc)` — Custom worker implementation with shutdown callback

### WORKER_ID Convention

- **Numbering**: Uses 1-based indexing (1, 2, 3, ...) for user-friendly logging
- **Appears in**: Worker process logs and environment variable
- **Example**: "Worker 1", "Worker 2", "Worker 3"
- **Design rationale**: WORKER_ID is user-facing; humans expect numbering to start at 1

### Option Struct Usage Examples

**Default behavior (backward compatible):**

```go
worker.RunMasterServer(8080, os.Args[0])
// Spawns 5 workers with no additional arguments
```

**Custom worker count:**

```go
worker.RunMasterServer(8080, os.Args[0], worker.Option{NumWorker: 4})
// Spawns 4 workers instead of default 5
```

**With both args and workers:**

```go
worker.RunMasterServer(8080, os.Args[0], worker.Option{
    NumWorker: 8,
    Args:      []string{"-verbose", "-timeout", "30s"},
})
// Spawns 8 workers, passes additional arguments
```

### Backward Compatibility

All Option-based changes are **fully backward compatible**:

- Existing code calling `RunMasterServer(port, cmd)` continues to work
- Default worker count remains 5
- Opt-in configuration through Option struct

### Commit Messages

- Use clear, descriptive titles (e.g., "Add graceful shutdown timeout to workers")
- Include brief explanation of **why** the change was needed
- Reference related issues if applicable
- Example:

  ```plaintext
  Improve master shutdown timeout handling

  Master process now waits 60s instead of 30s before force-killing
  workers, giving long-running requests more time to complete.

  Addresses: #15
  ```

**Note**: All commits require the Copilot co-author trailer:

```plaintext
Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>
```

### Code Review Expectations

- **Package changes** (pkg/worker/): Require review for API stability
- **Experiments** (cmd/v*/main.go): Require review for clarity and correctness
- **Documentation**: Markdown must pass markdownlint (relaxed config in `.markdownlint.yml`)

### Testing

All contributions must pass tests with the race detector:

```bash
make test
```

For experiments that add new behavior, consider adding test cases to the package. Refer to existing test patterns in the codebase.

### Using GitNexus for Safe Changes

Before making significant changes to pkg/worker/, run impact analysis:

```bash
gitnexus_impact({target: "RunMasterServer", direction: "upstream"})
```

This shows all callers and helps you understand the change's scope. For detailed guidance, see CLAUDE.md.

## Common Contribution Types

### Adding a New Experiment

1. Create `cmd/v{N}/main.go` where N is the next version
2. Copy and modify from the previous version as a starting point
3. Document the experiment goal in a comment at the top
4. Update README.md with the new experiment
5. Run `make build` and `make test`
6. Commit with a clear message explaining the research goal

### Improving the Package API

1. Identify the public function(s) to change in `pkg/worker/`
2. Check impact with GitNexus
3. Update `pkg/worker/examples/main.go` to reflect the new API
4. Update inline documentation
5. Run `make test` to ensure no breakage
6. Update `.github/copilot-instructions.md` if the API changes significantly

### Fixing a Bug

1. Add a test case that reproduces the bug (if not already tested)
2. Fix the bug
3. Verify the test passes: `make test`
4. Note the bug and fix in the commit message

### Improving Documentation

1. Edit README.md, CONTRIBUTING.md, or `.github/copilot-instructions.md`
2. Check formatting with markdownlint (if available in your editor)
3. Commit with a message like "docs: improve systemd integration explanation"

## Code Style Guide

### Go

- **Naming**: Use camelCase for variables/functions, PascalCase for exported types
- **Formatting**: Run `gofmt` (automatic via build)
- **Error handling**: Always check and log errors; don't silently ignore
- **Comments**: Export all exported functions with comment summary

  ```go
  // RunMasterServer starts a master process listening on the given port.
  func RunMasterServer(port int, command string, args ...string) { ... }
  ```

### Markdown

- Follow `.markdownlint.yml` relaxed rules
- Use consistent heading hierarchy (# → ## → ###)
- Code blocks should specify language: ` ```bash ` or ` ```go `
- Keep line length reasonable for readability (no hard limit enforced)

## Troubleshooting

### Tests Fail with Race Detector Warnings

This usually indicates unsafe concurrent access. If you modified code touching goroutines, channels, or shared state:

- Review the change for synchronization issues
- Use `sync.Mutex` or channels to protect shared data
- See `pkg/worker/master.go` for examples of safe signal handling

### Makefile Targets Not Working

Ensure you're in the repo root and have run:

```bash
make fetch-module  # First time or after go.mod changes
```

### Build Succeeds but Tests Fail

The binary and tests may use different code paths. Always run:

```bash
make clean && make test
```

## Questions?

- Check README.md and `.github/copilot-instructions.md` for architecture overview
- Read CLAUDE.md for GitNexus guidance on code navigation
- Review existing experiments in `cmd/v*/main.go` for patterns
- Examine `pkg/worker/examples/main.go` for API usage

## Summary

1. **Understand the lab structure**: Experiments are unstable, package is stable
2. **Run tests before committing**: `make test`
3. **Document your changes**: Clear commit messages and code comments
4. **Use GitNexus for package changes**: Check impact before modifying exported symbols
5. **Follow Go conventions**: Format, naming, error handling
6. **Respect backward compatibility**: In pkg/worker/ only

Thank you for contributing! 🙏
