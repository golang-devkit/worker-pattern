package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type RequestResult struct {
	Duration   time.Duration
	Error      error
	StatusCode int
}

type StressTestConfig struct {
	URL            string
	Concurrency    int
	TotalRequests  int
	Timeout        time.Duration
	RampUpDuration time.Duration
}

type StressTestStats struct {
	TotalRequests int64
	SuccessCount  int64
	ErrorCount    int64
	TotalDuration time.Duration
	Latencies     []time.Duration
	MinLatency    time.Duration
	MaxLatency    time.Duration
	AvgLatency    time.Duration
	P50Latency    time.Duration
	P95Latency    time.Duration
	P99Latency    time.Duration
	Throughput    float64
}

func main() {
	var (
		url         = flag.String("url", "http://27.71.229.15:3000/heavy", "Target URL")
		concurrency = flag.Int("concurrency", 1, "Number of concurrent requests")
		totalReqs   = flag.Int("requests", 10, "Total number of requests")
		timeout     = flag.Duration("timeout", 60*time.Second, "Request timeout")
		rampUp      = flag.Duration("rampup", 0, "Ramp-up duration (0 = no ramp-up)")
	)
	flag.Parse()

	config := StressTestConfig{
		URL:            *url,
		Concurrency:    *concurrency,
		TotalRequests:  *totalReqs,
		Timeout:        *timeout,
		RampUpDuration: *rampUp,
	}

	fmt.Println("================================================================================")
	fmt.Println("STRESS TEST: Worker Pattern Heavy Endpoint")
	fmt.Println("================================================================================")
	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  URL:              %s\n", config.URL)
	fmt.Printf("  Concurrency:      %d\n", config.Concurrency)
	fmt.Printf("  Total Requests:   %d\n", config.TotalRequests)
	fmt.Printf("  Timeout:          %v\n", config.Timeout)
	fmt.Printf("  Ramp-up Duration: %v\n", config.RampUpDuration)
	fmt.Printf("\n================================================================================\n")

	stats := runStressTest(config)
	printStats(stats)
}

func runStressTest(config StressTestConfig) StressTestStats {
	client := &http.Client{
		Timeout: config.Timeout,
	}

	// Create channels and synchronization
	requestChan := make(chan int, config.TotalRequests)
	resultChan := make(chan RequestResult, config.TotalRequests)
	var wg sync.WaitGroup

	// Fill request channel
	for i := 0; i < config.TotalRequests; i++ {
		requestChan <- i
	}
	close(requestChan)

	// Track stats
	var (
		successCount   int64
		errorCount     int64
		latencies      []time.Duration
		latenciesMutex sync.Mutex
		startTime      = time.Now()
	)

	// Calculate ramp-up increment
	rampUpIncrement := time.Duration(0)
	if config.RampUpDuration > 0 && config.Concurrency > 1 {
		rampUpIncrement = config.RampUpDuration / time.Duration(config.Concurrency)
	}

	// Start workers
	for i := 0; i < config.Concurrency; i++ {
		if config.RampUpDuration > 0 {
			time.Sleep(rampUpIncrement)
		}

		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for req := range requestChan {
				result := makeRequest(client, config.URL, req, workerID)
				resultChan <- result

				if result.Error == nil {
					atomic.AddInt64(&successCount, 1)
				} else {
					atomic.AddInt64(&errorCount, 1)
				}

				latenciesMutex.Lock()
				latencies = append(latencies, result.Duration)
				latenciesMutex.Unlock()
			}
		}(i)
	}

	// Wait for all workers to finish
	wg.Wait()
	close(resultChan)

	// Collect all results
	for range resultChan {
	}

	totalDuration := time.Since(startTime)

	// Calculate stats
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	stats := StressTestStats{
		TotalRequests: int64(config.TotalRequests),
		SuccessCount:  atomic.LoadInt64(&successCount),
		ErrorCount:    atomic.LoadInt64(&errorCount),
		TotalDuration: totalDuration,
		Latencies:     latencies,
	}

	// Calculate latency percentiles
	if len(latencies) > 0 {
		stats.MinLatency = latencies[0]
		stats.MaxLatency = latencies[len(latencies)-1]

		var sum time.Duration
		for _, lat := range latencies {
			sum += lat
		}
		stats.AvgLatency = time.Duration(int64(sum) / int64(len(latencies)))

		stats.P50Latency = percentile(latencies, 50)
		stats.P95Latency = percentile(latencies, 95)
		stats.P99Latency = percentile(latencies, 99)

		stats.Throughput = float64(stats.SuccessCount) / stats.TotalDuration.Seconds()
	}

	return stats
}

func makeRequest(client *http.Client, url string, reqNum, workerID int) RequestResult {
	start := time.Now()

	resp, err := client.Get(url)
	duration := time.Since(start)

	if err != nil {
		return RequestResult{
			Duration:   duration,
			Error:      err,
			StatusCode: 0,
		}
	}
	defer resp.Body.Close()

	// Consume response body to ensure full request completion
	_, _ = io.ReadAll(resp.Body)

	return RequestResult{
		Duration:   duration,
		Error:      nil,
		StatusCode: resp.StatusCode,
	}
}

func percentile(latencies []time.Duration, p int) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	index := int(math.Ceil(float64(len(latencies))*float64(p)/100.0)) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(latencies) {
		index = len(latencies) - 1
	}
	return latencies[index]
}

func printStats(stats StressTestStats) {
	fmt.Println("\nRESULTS")
	fmt.Println("================================================================================")

	// Summary
	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Total Requests:     %d\n", stats.TotalRequests)
	fmt.Printf("  Successful:         %d (%.2f%%)\n", stats.SuccessCount,
		float64(stats.SuccessCount)/float64(stats.TotalRequests)*100)
	fmt.Printf("  Errors:             %d (%.2f%%)\n", stats.ErrorCount,
		float64(stats.ErrorCount)/float64(stats.TotalRequests)*100)
	fmt.Printf("  Total Duration:     %.2fs\n", stats.TotalDuration.Seconds())
	fmt.Printf("  Throughput:         %.2f req/s\n", stats.Throughput)

	// Latency stats
	fmt.Printf("\nLatency (ms):\n")
	fmt.Printf("  Min:                %.2f\n", stats.MinLatency.Seconds()*1000)
	fmt.Printf("  Max:                %.2f\n", stats.MaxLatency.Seconds()*1000)
	fmt.Printf("  Avg:                %.2f\n", stats.AvgLatency.Seconds()*1000)
	fmt.Printf("  P50 (Median):       %.2f\n", stats.P50Latency.Seconds()*1000)
	fmt.Printf("  P95:                %.2f\n", stats.P95Latency.Seconds()*1000)
	fmt.Printf("  P99:                %.2f\n", stats.P99Latency.Seconds()*1000)

	fmt.Println("\n================================================================================")
}
