// pkg/stats/stats.go
package stats

import (
    "fmt"
    "sync"
    "sync/atomic"
    "time"
)

type Collector struct {
    totalRequests   uint64
    successReqs     uint64
    failedReqs      uint64
    activeWorkers   int32
    totalLatency    uint64
    statusCodes     map[int]uint64
    errors          map[string]uint64
    slowlorisKeeps  uint64
    mu              sync.RWMutex
    startTime       time.Time
    stopDisplay     chan bool
}

func NewCollector() *Collector {
    return &Collector{
        statusCodes: make(map[int]uint64),
        errors:      make(map[string]uint64),
        startTime:   time.Now(),
        stopDisplay: make(chan bool),
    }
}

func (c *Collector) IncrementActive() {
    atomic.AddInt32(&c.activeWorkers, 1)
}

func (c *Collector) DecrementActive() {
    atomic.AddInt32(&c.activeWorkers, -1)
}

func (c *Collector) RecordSuccess(statusCode int, latency time.Duration) {
    atomic.AddUint64(&c.totalRequests, 1)
    atomic.AddUint64(&c.successReqs, 1)
    atomic.AddUint64(&c.totalLatency, uint64(latency.Microseconds()))
    
    c.mu.Lock()
    c.statusCodes[statusCode]++
    c.mu.Unlock()
}

func (c *Collector) RecordFailure(errMsg string) {
    atomic.AddUint64(&c.totalRequests, 1)
    atomic.AddUint64(&c.failedReqs, 1)
    
    c.mu.Lock()
    c.errors[errMsg]++
    c.mu.Unlock()
}

func (c *Collector) RecordSlowlorisKeepalive() {
    atomic.AddUint64(&c.slowlorisKeeps, 1)
}

func (c *Collector) Start() {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-c.stopDisplay:
            return
        case <-ticker.C:
            c.display()
        }
    }
}

func (c *Collector) display() {
    elapsed := time.Since(c.startTime)
    total := atomic.LoadUint64(&c.totalRequests)
    success := atomic.LoadUint64(&c.successReqs)
    failed := atomic.LoadUint64(&c.failedReqs)
    active := atomic.LoadInt32(&c.activeWorkers)
    rps := float64(total) / elapsed.Seconds()
    
    var avgLatency float64
    if success > 0 {
        avgLatency = float64(atomic.LoadUint64(&c.totalLatency)/success) / 1000
    }
    
    fmt.Printf("\r[%s] Req: %d | RPS: %.0f | Success: %d | Fail: %d | Active: %d | Latency: %.2fms",
        elapsed.Round(time.Second).String(),
        total,
        rps,
        success,
        failed,
        active,
        avgLatency,
    )
}

func (c *Collector) PrintFinal() {
    close(c.stopDisplay)
    time.Sleep(100 * time.Millisecond)
    
    elapsed := time.Since(c.startTime)
    total := atomic.LoadUint64(&c.totalRequests)
    success := atomic.LoadUint64(&c.successReqs)
    failed := atomic.LoadUint64(&c.failedReqs)
    rps := float64(total) / elapsed.Seconds()
    keeps := atomic.LoadUint64(&c.slowlorisKeeps)
    
    fmt.Printf("\n\n═══════════════════════════════════════════════\n")
    fmt.Printf("FINAL STATISTICS\n")
    fmt.Printf("═══════════════════════════════════════════════\n")
    fmt.Printf("Total Requests:     %d\n", total)
    fmt.Printf("Successful:         %d (%.2f%%)\n", success, float64(success)/float64(total)*100)
    fmt.Printf("Failed:             %d (%.2f%%)\n", failed, float64(failed)/float64(total)*100)
    fmt.Printf("Requests/sec:       %.2f\n", rps)
    fmt.Printf("Duration:           %s\n", elapsed.Round(time.Millisecond))
    if keeps > 0 {
        fmt.Printf("Slowloris Keepalives: %d\n", keeps)
    }
    
    fmt.Printf("\nStatus Code Breakdown:\n")
    c.mu.RLock()
    for code, count := range c.statusCodes {
        fmt.Printf("  %d: %d\n", code, count)
    }
    c.mu.RUnlock()
    
    if len(c.errors) > 0 {
        fmt.Printf("\nError Breakdown:\n")
        c.mu.RLock()
        for err, count := range c.errors {
            if len(err) > 50 {
                err = err[:50] + "..."
            }
            fmt.Printf("  %s: %d\n", err, count)
        }
        c.mu.RUnlock()
    }
    fmt.Printf("═══════════════════════════════════════════════\n")
}
