# Stress Test Tool for Worker Pattern

A comprehensive HTTP stress testing tool designed to measure the
performance characteristics of the Worker Pattern `/heavy` endpoint under
various load conditions.

## Overview

This tool simulates concurrent HTTP requests to stress test the deployed
worker pattern application and collect detailed performance metrics
including:

- Request success/failure rates
- Latency percentiles (P50, P95, P99)
- Throughput (requests/second)
- Min/Max/Avg response times

## Features

✅ **Configurable Concurrency** - Control number of concurrent request streams
✅ **Flexible Load** - Specify total requests and ramp-up patterns
✅ **Detailed Metrics** - Latency percentiles, throughput, error rates
✅ **Timeout Protection** - Configurable per-request timeouts
✅ **Zero Dependencies** - Uses only Go stdlib

## Building

```bash
cd /Users/admin/go/src/lab_WorkerPattern
make build
# Binary: ./bin/stresstest
```

Or build standalone:

```bash
cd cmd/stresstest
go build -o ../../bin/stresstest main.go
```

## Usage

### Basic Test (1 worker, 10 requests)

```bash
./bin/stresstest
```

Default target: `http://27.71.229.15:3000/heavy`

### Custom URL

```bash
./bin/stresstest -url http://localhost:3000/heavy
```

### Load Testing with Concurrency

```bash
# 5 concurrent workers, 100 total requests
./bin/stresstest -concurrency=5 -requests=100

# 20 concurrent workers, 1000 total requests
./bin/stresstest -concurrency=20 -requests=1000
```

### Ramp-up Testing (gradual load increase)

```bash
# Start 10 workers over 10 seconds (1 per second)
./bin/stresstest -concurrency=10 -requests=100 -rampup=10s
```

### Custom Timeout

```bash
# 2 minute timeout per request
./bin/stresstest -timeout=120s -requests=50
```

### Combined Example

```bash
# Test with 8 concurrent requests over 8 seconds (1/sec ramp-up), 200 total requests
./bin/stresstest \
  -url http://27.71.229.15:3000/heavy \
  -concurrency=8 \
  -requests=200 \
  -rampup=8s \
  -timeout=60s
```

## Command-Line Flags

| Flag | Default | Description |
| --- | --- | --- |
| `-url` | `http://27.71.229.15:3000/heavy` | Target URL |
| `-concurrency` | `1` | Concurrent request streams |
| `-requests` | `10` | Total HTTP requests |
| `-timeout` | `60s` | Timeout per request |
| `-rampup` | `0` | Ramp-up duration |

## Output Format

The tool displays:

1. **Configuration** - Your test parameters
2. **Results** - Aggregate statistics:
   - Success/failure counts and percentages
   - Total duration
   - Throughput (requests/second)
   - Latency metrics (min, max, avg, percentiles)

Example output:

```text
================================================================================
STRESS TEST: Worker Pattern Heavy Endpoint
================================================================================

Configuration:
  URL:              http://27.71.229.15:3000/heavy
  Concurrency:      1
  Total Requests:   10
  Timeout:          1m0s
  Ramp-up Duration: 0s

================================================================================

RESULTS
================================================================================

Summary:
  Total Requests:     10
  Successful:         10 (100.00%)
  Errors:             0 (0.00%)
  Total Duration:     523.42s
  Throughput:         0.02 req/s

Latency (ms):
  Min:                52123.45
  Max:                52456.78
  Avg:                52289.56
  P50 (Median):       52300.12
  P95:                52450.00
  P99:                52456.78

================================================================================
```

## Interpreting Results

### Throughput

- **Formula**: `Successful Requests / Total Duration`
- **For `/heavy` endpoint**: Expected ~0.02-0.05 req/s (each request does 100M iterations)
- **Interpretation**:
  - Single worker: limited by CPU-bound work
  - Multiple workers: should scale approximately with worker count

### Latency Percentiles

- **P50 (Median)**: 50% of requests finish within this time (typical response)
- **P95**: 95% of requests finish within this time (good SLA target)
- **P99**: 99% of requests finish within this time (worst-case acceptable)

### Example Analysis (1 worker)

```text
P50: 52.3s  → Half of requests complete in ~52 seconds
P95: 52.4s  → 95% complete within 52.4 seconds  
P99: 52.5s  → Only 1% of requests exceed 52.5 seconds
```

This indicates consistent response times with little variation.

## Test Scenarios

### Scenario 1: Single Worker Baseline (1 worker)

```bash
./bin/stresstest -concurrency=1 -requests=10
```

**Goal**: Establish baseline performance with one worker
**Expected**: ~10 seconds per request (100M iterations)

### Scenario 2: Scalability Test (Multiple workers)

```bash
# Test with different worker counts
for workers in 1 2 4 8; do
  echo "Testing with $workers workers..."
  ./bin/stresstest -concurrency=$workers -requests=100
done
```

**Goal**: Verify linear scaling with worker count
**Expected**: Throughput increases ~linearly with workers

### Scenario 3: Sustained Load

```bash
./bin/stresstest -concurrency=4 -requests=1000
```

**Goal**: Test stability under sustained load
**Expected**: Consistent latencies, no degradation

### Scenario 4: Ramp-up Behavior

```bash
./bin/stresstest -concurrency=8 -requests=200 -rampup=8s
```

**Goal**: Smooth load increase, observe startup overhead
**Expected**: Initial requests may be slower, then stabilize

## Performance Notes

### The `/heavy` Endpoint

The `/heavy` endpoint in the worker pattern does:

```go
for i := range 100000000 {  // 100 million iterations
    sum += i
}
```

This is intentionally CPU-intensive to:

- Demonstrate multi-worker benefit (parallel CPU work)
- Measure kernel scheduling efficiency
- Show fair connection distribution

### Single Worker Characteristics

With 1 worker:

- **Expected Latency**: ~50-60 seconds per request
- **Expected Throughput**: ~0.017-0.02 req/s
- **Expected CPU**: ~95-100% utilization
- **Bottleneck**: CPU-bound computation on single core

### Multi-Worker Characteristics

With N workers:

- **Throughput**: Scales approximately linearly (up to number of CPU cores)
- **Latency**: Minimal change (requests still take ~50-60s each)
- **CPU**: Each worker gets dedicated core via `GOMAXPROCS(1)`
- **Benefit**: Can handle more concurrent requests

## Comparing Deployments

To compare performance across different deployments:

```bash
# Test local deployment
echo "Local (localhost:3000):"
./bin/stresstest -url http://localhost:3000/heavy -concurrency=1 -requests=5

# Test remote deployment  
echo "Remote (27.71.229.15:3000):"
./bin/stresstest -url http://27.71.229.15:3000/heavy -concurrency=1 -requests=5

# Compare throughput and latencies
```

## Troubleshooting

### All requests timeout

- Increase `-timeout` flag
- Verify endpoint is accessible: `curl http://27.71.229.15:3000/heavy`
- Check network connectivity and firewall rules

### High error rate

- Reduce concurrency (server may be overloaded)
- Increase request timeout
- Check server logs for issues

### Unexpected latency

- For `/heavy` endpoint, expect ~50-60 seconds per request
- Higher latency may indicate network/server issues
- Compare with known baseline values

## Code Structure

- **`main()`** - CLI flag parsing and startup
- **`runStressTest()`** - Orchestrates worker goroutines and request distribution
- **`makeRequest()`** - Individual HTTP request execution
- **`percentile()`** - Calculates latency percentiles
- **`printStats()`** - Formats and displays results

## Implementation Details

### Concurrency Model

- Each concurrent stream runs in its own goroutine
- Requests distributed via unbuffered channel for fair scheduling
- Atomic counters for thread-safe statistics

### Ramp-up Strategy

- Delays worker startup by `rampup_duration / concurrency`
- Provides gradual load increase for observing startup behavior
- Useful for detecting initialization overhead

### Accuracy

- Uses `time.Now()` for nanosecond-precision timing
- Sorts latencies for accurate percentile calculation
- Consumes full response body for complete request measurement

## Future Enhancements

Potential additions:

- CSV output for data analysis
- Real-time progress display
- Histogram visualization
- Custom request headers
- POST/PUT/DELETE method support
- Response body validation
- Request rate limiting (transactions/sec target)

## Related Tools

- `wrk` - Popular HTTP benchmarking tool (more features)
- `ab` - Apache Bench (simpler, UNIX standard)
- `hey` - Modern Go-based HTTP load generator
- `vegeta` - Attack-oriented load testing tool
