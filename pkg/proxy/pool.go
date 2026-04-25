// pkg/proxy/pool.go
package proxy

import (
    "sync"
    "time"
)

type ProxyPool struct {
    proxies   []string
    available chan string
    mu        sync.RWMutex
    blacklist map[string]time.Time
}

func NewProxyPool(proxies []string) *ProxyPool {
    pool := &ProxyPool{
        proxies:   proxies,
        available: make(chan string, len(proxies)),
        blacklist: make(map[string]time.Time),
    }
    
    for _, p := range proxies {
        pool.available <- p
    }
    
    return pool
}

func (p *ProxyPool) Get() string {
    select {
    case proxy := <-p.available:
        p.mu.RLock()
        if banTime, ok := p.blacklist[proxy]; ok && time.Since(banTime) < 30*time.Second {
            p.mu.RUnlock()
            time.Sleep(100 * time.Millisecond)
            return p.Get()
        }
        p.mu.RUnlock()
        return proxy
    default:
        return ""
    }
}

func (p *ProxyPool) Release(proxy string, success bool) {
    if !success {
        p.mu.Lock()
        p.blacklist[proxy] = time.Now()
        p.mu.Unlock()
        time.Sleep(1 * time.Second)
    }
    p.available <- proxy
}

func (p *ProxyPool) Size() int {
    return len(p.proxies)
}
