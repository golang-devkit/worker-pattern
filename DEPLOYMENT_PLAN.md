# Deployment Plan: pkg/worker Improvements

**Status**: Ready for Implementation  
**Total Tasks**: 12  
**Estimated Time**: ~51 minutes  
**Target Completion**: April 3, 2026

---

## Executive Summary

This deployment plan implements the 4 agreed improvements to `pkg/worker/` package:

1. **Priority 1 (HIGH)**: Configurable worker count via Option struct (5 subtasks)
2. **Priority 2 (MEDIUM)**: Error handling on process signals (2 subtasks)
3. **Priority 3 (MEDIUM)**: Explicit shutdown ordering (2 subtasks)
4. **Priority 4 (LOW)**: Documentation & examples (2 subtasks + 1 validation)

---

## Task Breakdown

### PHASE 1: Option Struct Implementation (25 minutes)

#### Task DEPLOY-1: Define Option struct
- **File**: `pkg/worker/master.go`
- **Lines**: Top of file (after imports)
- **What**: Add type definition
- **Time**: 5 min
- **Code**:
```go
type Option struct {
    Args      []string
    NumWorker int32
}
```

#### Task DEPLOY-2: Update RunMasterServer signature
- **File**: `pkg/worker/master.go`
- **Lines**: 14 (current)
- **What**: Change function signature
- **Time**: 5 min
- **From**:
```go
func RunMasterServer(port int, command string, args ...string)
```
- **To**:
```go
func RunMasterServer(port int, command string, opts ...Option)
```
- **Also**: Update call to RunMasterWith to pass opts

#### Task DEPLOY-3: Update RunMasterWith signature
- **File**: `pkg/worker/master.go`
- **Lines**: 33 (current)
- **What**: Change function signature
- **Time**: 3 min
- **From**:
```go
func RunMasterWith(extraFiles []*os.File, command string, args ...string)
```
- **To**:
```go
func RunMasterWith(extraFiles []*os.File, command string, opts ...Option)
```

#### Task DEPLOY-4: Implement Option handling logic
- **File**: `pkg/worker/master.go`
- **Lines**: 35-39 (new logic)
- **What**: Extract values from opts with defaults
- **Time**: 5 min
- **Code**:
```go
// Determine worker count and arguments
numWorkers := 5  // default
args := []string{}

if len(opts) > 0 {
    if opts[0].NumWorker > 0 {
        numWorkers = int(opts[0].NumWorker)
    }
    if opts[0].Args != nil {
        args = opts[0].Args
    }
}

// Update line that calls SpawnWorker
sp := SpawnWorker(numWorkers, extraFiles, command, args...)
```

#### Task DEPLOY-5: Update examples/main.go
- **File**: `pkg/worker/examples/main.go`
- **Lines**: 18 (main())
- **What**: Show both old and new API usage
- **Time**: 5 min
- **Code**:
```go
// Example 1: Backward compatible (default 5 workers)
worker.RunMasterServer(8080, os.Args[0])

// Example 2: Custom worker count
worker.RunMasterServer(8080, os.Args[0], worker.Option{NumWorker: 4})

// Example 3: Both args and workers
worker.RunMasterServer(8080, os.Args[0], worker.Option{
    NumWorker: 8,
    Args:      []string{"-verbose"},
})
```

---

### PHASE 2: Error Handling (8 minutes)

#### Task DEPLOY-6: Add error handling to Signal in spawn.go
- **File**: `pkg/worker/spawn.go`
- **Lines**: 45 (current)
- **What**: Check Signal() error and log
- **Time**: 5 min
- **From**:
```go
worker.Process.Signal(syscall.SIGTERM)
```
- **To**:
```go
if err := worker.Process.Signal(syscall.SIGTERM); err != nil {
    log.Printf("Warning: failed to signal worker %d (PID %d): %v", 
        i, worker.Process.Pid, err)
}
```

#### Task DEPLOY-7: Add error handling to Signal in cmd/v1
- **File**: `cmd/v1/main.go`
- **Lines**: 105 (current)
- **What**: Check Signal() error and log
- **Time**: 3 min
- **Same pattern as DEPLOY-6**

---

### PHASE 3: Shutdown Ordering (5 minutes)

#### Task DEPLOY-8: Remove defer from master.go
- **File**: `pkg/worker/master.go`
- **Lines**: 37 (current)
- **What**: Delete the defer sp.Stop() line
- **Time**: 2 min
- **Remove**:
```go
defer sp.Stop()
```

#### Task DEPLOY-9: Make shutdown explicit
- **File**: `pkg/worker/master.go`
- **Lines**: After line 66 (after sdNotify("STOPPING=1"))
- **What**: Call sp.Stop() directly in signal handler
- **Time**: 3 min
- **Add**:
```go
// Explicit shutdown (not deferred)
sp.Stop()
os.Remove(pidFile)
log.Println("Master process exiting")
```
- **Remove**: Old cleanup code (replaced above)

---

### PHASE 4: Documentation (10 minutes)

#### Task DEPLOY-10: Update CONTRIBUTING.md
- **File**: `CONTRIBUTING.md`
- **Section**: "Key Conventions and Patterns"
- **What**: Document WORKER_ID is 1-based (user-friendly)
- **Time**: 5 min
- **Add**:
```markdown
### Public API (pkg/worker/)
- **WORKER_ID Numbering**: Uses 1-based indexing (1, 2, 3...) for user-friendly logging
- `RunMasterServer(port, command, opts ...Option)` — Start master with configurable workers
- `RunMasterWith(files, command, opts ...Option)` — Advanced: custom file descriptors
- `RegisterWorker(executeFunc)` — Custom worker implementation with shutdown callback

### Option Struct
- **NumWorker** (int32): Number of worker processes (default: 5)
- **Args** ([]string): Additional arguments to pass to workers
```

#### Task DEPLOY-11: Add Option struct examples
- **File**: `CONTRIBUTING.md`
- **Section**: New "API Usage Examples" subsection
- **What**: Show practical usage patterns
- **Time**: 5 min
- **Add**:
```markdown
### Option Struct Examples

**Default behavior (backward compatible):**
```go
worker.RunMasterServer(8080, os.Args[0])
```

**Custom worker count:**
```go
worker.RunMasterServer(8080, os.Args[0], worker.Option{NumWorker: 4})
```

**With both args and workers:**
```go
worker.RunMasterServer(8080, os.Args[0], worker.Option{
    NumWorker: 8,
    Args:      []string{"-verbose", "-timeout", "30s"},
})
```
```

---

### PHASE 5: Validation (10 minutes)

#### Task DEPLOY-12: Run tests and validation
- **Command**: `make test`
- **What**: Verify all changes compile and pass race detector
- **Time**: 10 min
- **Checklist**:
  - ✓ Code compiles without errors
  - ✓ All tests pass with race detector
  - ✓ WORKER_ID still 1-based (check logs)
  - ✓ Backward compatibility works (RunMasterServer(port, cmd))
  - ✓ New API works (RunMasterServer(port, cmd, Option{...}))
  - ✓ Error messages logged on signal failures
  - ✓ Shutdown sequence is explicit and logged

---

## Implementation Order

```
PHASE 1: Option Struct (25 min)
  └─ DEPLOY-1 → DEPLOY-2 → DEPLOY-3 → DEPLOY-4 → DEPLOY-5

PHASE 2: Error Handling (8 min)
  └─ DEPLOY-6 → DEPLOY-7

PHASE 3: Shutdown Ordering (5 min)
  └─ DEPLOY-8 → DEPLOY-9

PHASE 4: Documentation (10 min)
  └─ DEPLOY-10 → DEPLOY-11

PHASE 5: Validation (10 min)
  └─ DEPLOY-12
```

**Total Estimated Time**: ~51 minutes (with buffer: 1 hour)

---

## Rollback Plan

If any phase fails:

1. **PHASE 1 fails**: Revert master.go to previous version; Option struct incomplete
   - Risk: LOW (isolated to master.go)
   
2. **PHASE 2 fails**: Revert to original Signal() calls (no error handling)
   - Risk: LOW (backward compatible)
   
3. **PHASE 3 fails**: Restore defer sp.Stop(); explicit shutdown not working
   - Risk: LOW (less clear but functionally same)
   
4. **PHASE 4 fails**: CONTRIBUTING.md is documentation only
   - Risk: NONE (skip to PHASE 5)

Rollback command: `git checkout -- pkg/worker/ cmd/v1/` (after backing up changes)

---

## Success Criteria

✅ All 12 tasks completed  
✅ `make test` passes with race detector  
✅ No regressions in existing functionality  
✅ New Option struct API is backward compatible  
✅ Error messages logged for signal failures  
✅ Shutdown sequence is explicit and traceable  
✅ Documentation reflects all changes  
✅ No new compiler warnings or errors  

---

## Files Modified

| File | Changes |
|------|---------|
| `pkg/worker/master.go` | Option struct, signatures, logic, shutdown |
| `pkg/worker/spawn.go` | Error handling on Signal() |
| `pkg/worker/examples/main.go` | Usage examples |
| `cmd/v1/main.go` | Error handling on Signal() |
| `CONTRIBUTING.md` | Documentation updates |

---

## Commit Strategy

After completion, commit with message:

```
feat(pkg/worker): Add Option struct for configurable worker count

- Add Option struct with NumWorker and Args fields
- Make RunMasterServer() and RunMasterWith() accept opts ...Option
- Default worker count to 5 (configurable per deployment)
- Add error handling for process.Signal() calls
- Make shutdown sequence explicit (remove defer)
- Verify backward compatibility with existing API
- Update documentation and examples

All tests passing. Backward compatible with existing code.

Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>
```

---

## Tracking with task-master-ai

To track this with task-master-ai (optional):

```bash
# Create parent task
task-master create "Implement pkg/worker improvements" \
  --description "Deploy 4 agreed improvements: Option struct, error handling, shutdown ordering, docs" \
  --priority high \
  --due "2026-04-03"

# Create subtasks (auto-generate from this plan)
task-master create "Define Option struct" --parent=... --assignee=assistant
task-master create "Update RunMasterServer" --parent=... --assignee=assistant
# ... etc for all 12 tasks
```

---

## Next Steps

1. Review this deployment plan
2. Approve implementation order
3. Execute PHASE 1-5 in sequence
4. Run validation (PHASE 5)
5. Commit changes with proper message
6. (Optional) Close corresponding GitHub issues
