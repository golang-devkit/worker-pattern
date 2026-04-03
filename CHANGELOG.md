# Changelog

<!-- markdownlint-disable MD013 MD025 MD033 MD060 -->

All notable changes to this project will be documented in this file.

## [v1.0.1] - 2026-04-03

**Full Changelog**: [Initial Release](https://github.com/golang-devkit/worker-pattern)

### 🔧 Worker Pattern Library — Master-Worker Socket Sharing Architecture (New Feature)

#### 📝 New Package — `pkg/worker`

##### Complete implementation of the Unix Systemd Worker Pattern with socket sharing

- **New Package**: `github.com/golang-devkit/worker-pattern/pkg/worker`
  - Stable, production-ready implementation extracted from research
  - Complete socket sharing architecture for multi-worker load distribution

#### 🎯 Implementation Details — v1.0.1 (Worker Pattern Library)

**Exported Functions:**

| Function | Purpose | Type |
|----------|---------|------|
| `RunMasterServer(port int, command string, opts ...Option)` | Starts master process with shared socket listener | New |
| `RunMasterWith(files []*os.File, command string, opts ...Option)` | Advanced master setup with pre-configured file descriptors | New |
| `SpawnWorker(n int, extraFiles []*os.File, name string, arg ...string) *Spawn` | Spawns worker processes with inherited file descriptors | New |
| `RegisterWorker(execute func(...))` | Registers worker process callback for socket handling | New |
| `sdNotify(state string) error` | Sends systemd readiness notifications | New |

**Core Types:**

| Type | Description | Impact |
|------|-------------|--------|
| `Option` struct | Configuration for master server (NumWorker, Args) | New |
| `Spawn` struct | Worker process lifecycle manager | New |

**Files Created (5 files):**

| File | Lines | Type | Impact |
|------|-------|------|--------|
| `pkg/worker/worker.go` | 51 | Worker lifecycle | Non-breaking |
| `pkg/worker/master.go` | 130+ | Master server setup | Non-breaking |
| `pkg/worker/spawn.go` | 55+ | Process spawning | Non-breaking |
| `pkg/worker/notify.go` | 20+ | Systemd integration | Non-breaking |
| `pkg/worker/examples/main.go` | 85+ | Example usage | Non-breaking |

#### ✅ Implementation Benefits — v1.0.1

- **Efficient Multi-Worker Architecture**: Master process creates a shared TCP socket and passes file descriptors to worker processes, enabling true load distribution without the need for ports or inter-process communication
- **Graceful Shutdown Support**: Workers handle SIGTERM/SIGINT signals with configurable shutdown callbacks and 30-second timeout for clean connection draining
- **Systemd Integration Ready**: Built-in `sdNotify()` support for systemd Type=notify services, enabling health monitoring and proper service lifecycle management
- **Option-Based Configuration API**: Flexible, builder-pattern style configuration without breaking API compatibility in future releases
- **CPU Core Affinity**: Each worker is limited to a single CPU core using `runtime.GOMAXPROCS(1)` to prevent over-subscription
- **Production-Ready Deployments**: Designed specifically for systemd unit file deployment with excellent control over resource usage and restart behavior

#### 📋 Additional Changes

- Added `LICENSE` file (MIT) for open-source distribution
- Added module `go.mod` files for `pkg/worker` as a standalone, importable library
- Updated `README.md` with architecture documentation and clarity on research/package separation

---
