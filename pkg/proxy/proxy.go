// pkg/proxy/proxy.go
package proxy

import (
    "fmt"
    "net"
    "net/url"
    "sync/atomic"
    "time"

    "golang.org/x/net/proxy"
)

type ProxyManager struct {
    proxies     []string
    currentIdx  uint64
    enabled     bool
    rotateEvery int
    requestCount uint64
}

func NewProxyManager(proxies []string, rotateEvery int) *ProxyManager {
    return &ProxyManager{
        proxies:     proxies,
        enabled:     len(proxies) > 0,
        rotateEvery: rotateEvery,
        currentIdx:  0,
    }
}

func (p *ProxyManager) GetProxy() (string, error) {
    if !p.enabled || len(p.proxies) == 0 {
        return "", nil
    }
    
    idx := atomic.LoadUint64(&p.currentIdx)
    if p.rotateEvery > 0 {
        reqCount := atomic.AddUint64(&p.requestCount, 1)
        if reqCount%uint64(p.rotateEvery) == 0 {
            idx = (idx + 1) % uint64(len(p.proxies))
            atomic.StoreUint64(&p.currentIdx, idx)
        }
    }
    
    return p.proxies[idx], nil
}

func (p *ProxyManager) CreateDialer(proxyURL string) (proxy.D
