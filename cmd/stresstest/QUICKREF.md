# Stress Test Tool - Quick Reference Card

## Build & Run

```bash
cd /Users/admin/go/src/lab_WorkerPattern

# Build
make stresstest

# Run default test
make stress

# Or run manually
./bin/stresstest [FLAGS]
```

## Common Commands

| Command | Purpose |
| --- | --- |
| `./bin/stresstest` | Default: 10 req to heavy endpoint |
| `./bin/stresstest -requests=1` | Single request baseline |
| `./bin/stresstest -requests=5 -timeout=120s` | 5 req, 2min timeout |

```bash
-url              Target URL (default: http://27.71.229.15:3000/heavy)
-concurrency      Concurrent streams (default: 1)
-requests         Total requests (default: 10)
-timeout          Per-request timeout (default: 60s)
-rampup           Ramp-up duration (default: 0 = immediate)
```

## Understanding Output

```text
RESULTS
================================================================================

Summary:
  Total Requests:     10          ← Total requests sent
  Successful:         10 (100%)   ← Success rate (100% = no errors)
  Errors:             0 (0%)      ← Failed requests
  Total Duration:     0.78s       ← Wall-clock time
  Throughput:         12.8 req/s  ← Successful / Duration

Latency (ms):
  Min:                62.70       ← Fastest response
  Max:                93.80       ← Slowest response
  Avg:                78.25       ← Average latency
  P50 (Median):       78.00       ← 50% complete within this time
  P95:                93.00       ← 95% complete within this time
  P99:                93.80       ← 99% complete within this time
```

## Test Scenarios

### Scenario 1: Single Request

```bash
./bin/stresstest -requests=1
```

### Scenario 2: Consistency (5 sequential)

```bash
./bin/stresstest -requests=5
```

### Scenario 3: Batch Load

```bash
./bin/stresstest -requests=50
```

### Scenario 4: With Ramp-up

```bash
./bin/stresstest -concurrency=8 -requests=100 -rampup=8s
```

### Scenario 5: Short Timeout (tests timeout handling)

```bash
./bin/stresstest -requests=5 -timeout=30s
```

## Key Metrics

**Throughput**:

- How many successful requests per second
- Formula: `Successful Requests / Total Duration`
- For `/heavy`: ~12-15 req/s (network speed, not server processing)

**Latency**:

- Response time from request to response received
- P50/P95/P99 = percentiles (what % of requests finish within time)
- For `/heavy`: ~63-94ms (network roundtrip)

**Success Rate**:

- % of requests that completed successfully
- Expected: 100% unless timeout too short
- If low: check network connectivity or increase timeout

## Troubleshooting

### Connection refused

```text
→ Server not running
→ Check: sudo systemctl status worker-pattern
```

### Timeout errors

```text
→ Timeout too short (30s) vs actual need (50-60s)
→ Use: -timeout=120s
```

### All errors

```text
→ Check network: ping 27.71.229.15
→ Check firewall: can access port 3000?
→ Check server: journalctl -u worker-pattern -f
```

## Performance Notes

The `/heavy` endpoint:

- Performs 100,000,000 loop iterations (CPU-bound)
- Takes ~50-60 seconds server-side
- Returns response quickly (~63-94ms latency = network time)

Single worker means:

- Can only serve one request at a time
- Requests queue on server
- Latency = network + server processing time

## Files

| File | Purpose |
| --- | --- |
| `main.go` | Tool source code |
| `README.md` | Complete documentation |
| `TESTING.md` | Test scenarios |
| `QUICKREF.md` | This file |

## Next Steps

1. Run baseline: `make stress`
2. Note the latency numbers
3. Scale to more workers
4. Run same test again
5. Compare throughput improvement
