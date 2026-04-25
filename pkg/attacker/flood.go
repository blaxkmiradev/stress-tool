// pkg/attacker/flood.go
package attacker

import (
    "sync"
    "time"
    
    "github.com/stresstest/layer7-flood/pkg/client"
    "github.com/stresstest/layer7-flood/pkg/config"
    "github.com/stresstest/layer7-flood/pkg/stats"
    "github.com/stresstest/layer7-flood/pkg/utils"
)

type Flood struct {
    cfg     *config.Config
    stats   *stats.Collector
    workers []*FloodWorker
    wg      sync.WaitGroup
    stopCh  chan struct{}
}

type FloodWorker struct {
    id       int
    client   *client.HttpClient
    cfg      *config.Config
    stats    *stats.Collector
    stopCh   <-chan struct{}
    rateTicker *time.Ticker
}

func NewFlood(cfg *config.Config, stats *stats.Collector) *Flood {
    return &Flood{
        cfg:    cfg,
        stats:  stats,
        workers: make([]*FloodWorker, 0),
        stopCh: make(chan struct{}),
    }
}

func (f *Flood) Start() {
    for i := 0; i < f.cfg.Attack.Threads; i++ {
        worker := &FloodWorker{
            id:     i,
            client: client.NewHttpClient(f.cfg),
            cfg:    f.cfg,
            stats:  f.stats,
            stopCh: f.stopCh,
        }
        
        ratePerThread := f.cfg.Attack.RatePerThread
        if ratePerThread > 0 {
            interval := time.Second / time.Duration(ratePerThread)
            worker.rateTicker = time.NewTicker(interval)
        }
        
        f.workers = append(f.workers, worker)
        f.wg.Add(1)
        go worker.run()
    }
    f.wg.Wait()
}

func (f *Flood) Stop() {
    close(f.stopCh)
    for _, w := range f.workers {
        if w.rateTicker != nil {
            w.rateTicker.Stop()
        }
        w.client.Close()
    }
    f.wg.Wait()
}

func (w *FloodWorker) run() {
    defer func() { w.stats.DecrementActive() }()
    w.stats.IncrementActive()
    
    for {
        select {
        case <-w.stopCh:
            return
        default:
            if w.rateTicker != nil {
                <-w.rateTicker.C
            }
            
            req := utils.BuildRequest(w.cfg, w.id)
            utils.CustomizeRequest(req, w.cfg)
            
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
