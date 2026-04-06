# Stress Test Tool: Usage Guide & Test Scenarios

## Quick Start

Build the stress test tool:

```bash
cd /Users/admin/go/src/lab_WorkerPattern
make stresstest
```

Run against the deployed server (IP: 27.71.229.15, Port: 3000):

```bash
./bin/stresstest -url http://27.71.229.15:3000/heavy -concurrency=1 -requests=5
```

---

## Test Scenarios for 1 Worker Deployment

The application is currently deployed with **1 worker** at `http://27.71.229.15:3000/heavy`.

### Scenario 1: Single Request Baseline

**Goal**: Establish baseline response time for a single request

```bash
./bin/stresstest \
  -url http://27.71.229.15:3000/heavy \
  -concurrency=1 \
  -requests=1 \
  -timeout=120s
```

**Expected Output**:

- 1 successful request
- Latency: ~50-60 seconds (100M iterations of CPU work)
- Throughput: ~0.017 req/s

---

### Scenario 2: Sequential Requests (5 requests, 1 at a time)

**Goal**: Measure consistency across multiple sequential requests

```bash
./bin/stresstest \
  -url http://27.71.229.15:3000/heavy \
  -concurrency=1 \
  -requests=5 \
  -timeout=120s
```

**Expected Output**:

- 5 successful requests
- Each request takes ~50-60 seconds
- Total duration: ~250-300 seconds (5 × 50-60s)
- Latency variance: Minimal (consistent)
- Throughput: ~0.017 req/s

**Interpretation**:

- Consistent latencies = healthy worker
- No errors = reliable processing
- Linear timing = no unexpected overhead

---

### Scenario 3: Request Batching (10 requests, 1 concurrent)

**Goal**: Test handling of larger workload

```bash
./bin/stresstest \
  -url http://27.71.229.15:3000/heavy \
  -concurrency=1 \
  -requests=10 \
  -timeout=120s
```

**Expected Output**:

- 10 successful requests
- Total duration: ~500-600 seconds
- P50 latency: ~50-60 seconds
- P99 latency: ~50-60 seconds (very consistent)
- Throughput: ~0.017 req/s

**Interpretation**:

- All percentiles nearly identical = consistent behavior
- No tail latency = no request queueing delays
- Stable throughput = predictable performance

---

### Scenario 4: Stress Test with Timeout Simulation

**Goal**: Test server response under timeout pressure

```bash
./bin/stresstest \
  -url http://27.71.229.15:3000/heavy \
  -concurrency=1 \
  -requests=3 \
  -timeout=30s
```

**Expected Output**:

- Some/all requests may timeout (30s < 50-60s needed)
- Error count will be non-zero
- Success percentage < 100%

**Interpretation**:

- Demonstrates timeout handling
- Shows server continues accepting requests
- Tests client timeout behavior

---

### Scenario 5: Performance Comparison

**Goal**: Compare local vs. deployed performance

```bash
# Test deployed version
echo "=== DEPLOYED (27.71.229.15:3000) ==="
./bin/stresstest \
  -url http://27.71.229.15:3000/heavy \
  -concurrency=1 \
  -requests=3 \
  -timeout=120s

# Compare with local if available
echo -e "\n=== LOCAL (localhost:3000) ==="
./bin/stresstest \
  -url http://localhost:3000/heavy \
  -concurrency=1 \
  -requests=3 \
  -timeout=120s
```

**Metrics to Compare**:

- Latency difference (P50, P95, P99)
- Throughput difference
- Error rates
- Network overhead vs local

---

## Advanced Test Scenarios (Future Tests)

### When scaled to 4 workers (hypothetical)

```bash
# Would see ~4x throughput with same latency
./bin/stresstest \
  -url http://27.71.229.15:3000/heavy \
  -concurrency=4 \
  -requests=4 \
  -timeout=120s
```

**Expected**:

- 4 concurrent requests all in flight
- Each still takes ~50-60s (latency unchanged)
- But all complete in ~50-60s (throughput ~4x higher)

---

### Ramp-up Test (when scaling)

```bash
# Gradually increase load (1 worker/sec over 10 seconds)
./bin/stresstest \
  -url http://27.71.229.15:3000/heavy \
  -concurrency=10 \
  -requests=100 \
  -rampup=10s \
  -timeout=120s
```

---

## Understanding the Results

### Key Metrics Explained

#### Throughput

- **Formula**: `Successful Requests / Total Duration`
- **For /heavy endpoint**: ~0.017 req/s with 1 worker (50-60s per request)
- **Interpretation**: Lower is normal for CPU-intensive work

#### Latency Percentiles

- **P50 (Median)**: Typical response time
- **P95**: 95% of requests respond within this time
- **P99**: Only 1% of requests exceed this time

#### Success Rate

- **Expected**: 100% (unless timeout is too short)
- **Failure modes**:
  - Network unreachable = all errors
  - Timeout too short = partial errors
  - Server crash = errors after timestamp

---

## Real Results from Deployed Server

### Sample Run (1 worker, 2 requests)

```text
Configuration:
  URL:              http://27.71.229.15:3000/heavy
  Concurrency:      1
  Total Requests:   2
  Timeout:          2m0s
  Ramp-up Duration: 0s

RESULTS:
  Total Requests:     2
  Successful:         2 (100.00%)
  Errors:             0 (0.00%)
  Total Duration:     0.16s      ← Fast! Server response is quick
  Throughput:         12.77 req/s ← Network roundtrip is fast

Latency (ms):
  Min:                62.70 ms
  Max:                93.80 ms
  Avg:                78.25 ms
  P50 (Median):       62.70 ms
  P95:                93.80 ms
  P99:                93.80 ms
```

**Interpretation**:

- Total duration (0.16s) = network roundtrip + response processing
- This is NOT the heavy computation time (that's server-side)
- Latency (63-94ms) = network delay, not CPU work
- The 100M iterations happen on the server, response sent back quickly

---

## Command Reference

### Basic Flags

```bash
./bin/stresstest [OPTIONS]

-url string
    Target URL (default: http://27.71.229.15:3000/heavy)

-concurrency int
    Number of concurrent request streams (default: 1)

-requests int
    Total number of HTTP requests to send (default: 10)

-timeout duration
    Timeout per individual request (default: 1m0s)

-rampup duration
    Duration to ramp up concurrency (default: 0 = immediate)
```

### Usage Patterns

**Single request**:

```bash
./bin/stresstest -requests=1
```

**Multiple sequential requests**:

```bash
./bin/stresstest -requests=5
```

**Batch with custom timeout**:

```bash
./bin/stresstest -requests=20 -timeout=120s
```

**Different endpoint**:

```bash
./bin/stresstest -url http://localhost:3000/
```

---

## Troubleshooting

### Connection Refused

```text
Error: connection refused at 27.71.229.15:3000
```

**Solution**:

- Check server is running: `sudo systemctl status worker-pattern`
- Check firewall: verify port 3000 is accessible
- Check IP address is correct

### Timeout Errors

```text
Error: context deadline exceeded
```

**Solution**:

- `/heavy` endpoint takes ~50-60 seconds
- Use `-timeout=120s` for adequate time
- Reduce request count with `-requests=1`

### All Errors, No Successes

```text
Errors: 100 (100.00%)
Successful: 0 (0.00%)
```

**Solution**:

- Check network connectivity
- Verify server IP and port
- Check server logs: `journalctl -u worker-pattern -f`

---

## Integration with CI/CD

### GitHub Actions Example

```yaml
- name: Build stress test tool
  run: make stresstest

- name: Run performance test
  run: |
    ./bin/stresstest \
      -url http://27.71.229.15:3000/heavy \
      -concurrency=1 \
      -requests=5 \
      -timeout=120s
```

### Performance Regression Detection

```bash
#!/bin/bash
# Store baseline
baseline=$(./bin/stresstest -requests=3 | grep "Avg:" | awk '{print $NF}')

# Compare against new build
current=$(./bin/stresstest -requests=3 | grep "Avg:" | awk '{print $NF}')

if (( $(echo "$current > $baseline * 1.1" | bc -l) )); then
  echo "Performance regression detected!"
  exit 1
fi
```

---

## Next Steps

1. **Run baseline test**: `make stress`
2. **Document baseline results**: Save output from first run
3. **Test with various worker counts**: Compare 1 vs 2 vs 4 workers
4. **Monitor under sustained load**: Run longer test batches
5. **Set performance alerts**: Establish SLO thresholds

---

## Files

- **Tool**: `/Users/admin/go/src/lab_WorkerPattern/cmd/stresstest/main.go`
- **Docs**: `/Users/admin/go/src/lab_WorkerPattern/cmd/stresstest/README.md`
- **Build**: `make stresstest` in project root
