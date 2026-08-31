// Command loadtest fires synthetic events at the ingestion API and reports the
// throughput and latency the API actually sustained.
//
// It exists so the numbers quoted in the README ("2,000+ events per second",
// "<10ms API response times") are something you can reproduce on demand rather
// than take on faith.
//
//	go run ./loadtest -n 20000 -c 100
//
// Flags:
//
//	-url    ingestion endpoint (default http://localhost:8080/api/v1/events)
//	-n      total events to send
//	-c      concurrent senders
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type event struct {
	UserID    string `json:"user_id"`
	Action    string `json:"action"`
	Timestamp string `json:"timestamp"`
}

func main() {
	url := flag.String("url", "http://localhost:8080/api/v1/events", "ingestion endpoint")
	total := flag.Int("n", 20000, "total number of events to send")
	concurrency := flag.Int("c", 100, "number of concurrent senders")
	flag.Parse()

	if *total <= 0 || *concurrency <= 0 {
		fmt.Fprintln(os.Stderr, "-n and -c must both be greater than zero")
		os.Exit(2)
	}

	// One shared client with a generous connection pool. Without raising
	// MaxIdleConnsPerHost, Go's default of 2 forces the load generator to open
	// and tear down a TCP connection per request, and you end up measuring
	// socket churn instead of the API.
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        *concurrency * 2,
			MaxIdleConnsPerHost: *concurrency * 2,
			MaxConnsPerHost:     *concurrency * 2,
			IdleConnTimeout:     30 * time.Second,
		},
	}

	// Pre-marshal the bodies so JSON encoding is not part of the measurement.
	actions := []string{"scan_product", "add_to_cart", "page_view", "button_click"}
	bodies := make([][]byte, *total)
	now := time.Now().UTC().Format(time.RFC3339)
	for i := range bodies {
		b, err := json.Marshal(event{
			UserID:    fmt.Sprintf("load-user-%d", i),
			Action:    actions[i%len(actions)],
			Timestamp: now,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to build payload: %v\n", err)
			os.Exit(1)
		}
		bodies[i] = b
	}

	fmt.Printf("Sending %d events to %s with %d concurrent senders...\n\n",
		*total, *url, *concurrency)

	var (
		accepted atomic.Int64
		failed   atomic.Int64
		mu       sync.Mutex
		samples  = make([]time.Duration, 0, *total)
	)

	jobs := make(chan []byte, *concurrency)
	var wg sync.WaitGroup

	for w := 0; w < *concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			local := make([]time.Duration, 0, *total / *concurrency+1)

			for body := range jobs {
				start := time.Now()
				resp, err := client.Post(*url, "application/json", bytes.NewReader(body))
				if err != nil {
					failed.Add(1)
					continue
				}
				// The body must be drained and closed or the connection cannot
				// be reused, and the pool above becomes pointless.
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				local = append(local, time.Since(start))

				if resp.StatusCode == http.StatusAccepted {
					accepted.Add(1)
				} else {
					failed.Add(1)
				}
			}

			mu.Lock()
			samples = append(samples, local...)
			mu.Unlock()
		}()
	}

	wallStart := time.Now()
	for _, b := range bodies {
		jobs <- b
	}
	close(jobs)
	wg.Wait()
	wall := time.Since(wallStart)

	report(wall, accepted.Load(), failed.Load(), samples)
}

func report(wall time.Duration, accepted, failed int64, samples []time.Duration) {
	sort.Slice(samples, func(i, j int) bool { return samples[i] < samples[j] })

	var sum time.Duration
	for _, d := range samples {
		sum += d
	}

	throughput := float64(accepted) / wall.Seconds()

	fmt.Println("── Results ─────────────────────────────────")
	fmt.Printf("  Accepted (202) : %d\n", accepted)
	fmt.Printf("  Failed         : %d\n", failed)
	fmt.Printf("  Wall clock     : %s\n", wall.Round(time.Millisecond))
	fmt.Printf("  Throughput     : %.0f events/sec\n", throughput)

	if len(samples) > 0 {
		fmt.Println()
		fmt.Println("── Per-request latency (client-side, incl. loopback) ──")
		fmt.Printf("  mean : %s\n", (sum / time.Duration(len(samples))).Round(time.Microsecond))
		fmt.Printf("  p50  : %s\n", percentile(samples, 0.50).Round(time.Microsecond))
		fmt.Printf("  p95  : %s\n", percentile(samples, 0.95).Round(time.Microsecond))
		fmt.Printf("  p99  : %s\n", percentile(samples, 0.99).Round(time.Microsecond))
		fmt.Printf("  max  : %s\n", samples[len(samples)-1].Round(time.Microsecond))
	}
	fmt.Println()
	fmt.Println("Note: throughput is what the API accepted onto the Kafka topic.")
	fmt.Println("Confirm the worker drained it with:")
	fmt.Println(`  docker exec postgres psql -U postgres -d analytics -c "SELECT count(*) FROM analytics_events;"`)
}

// percentile returns the value at the given rank in an already-sorted slice.
func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(p * float64(len(sorted)))
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
