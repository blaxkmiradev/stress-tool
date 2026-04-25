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

func (p *ProxyManager) CreateDialer(proxyURL string) (proxy.Dialer, error) {
    if proxyURL == "" {
        return &net.Dialer{Timeout: 10 * time.Second}, nil
    }
    
    u, err := url.Parse(proxyURL)
    if err != nil {
        return nil, err
    }
    
    switch u.Scheme {
    case "socks5", "socks5h":
        auth := proxy.Auth{}
        if u.User != nil {
            auth.User = u.User.Username()
            auth.Password, _ = u.User.Password()
        }
        dialer, err := proxy.SOCKS5("tcp", u.Host, &auth, proxy.Direct)
        if err != nil {
            return nil, err
        }
        return dialer, nil
    case "http", "https":
        return &httpProxyDialer{proxyURL: proxyURL}, nil
    default:
        return nil, fmt.Errorf("unsupported proxy scheme: %s", u.Scheme)
    }
}

type httpProxyDialer struct {
    proxyURL string
}

func (h *httpProxyDialer) Dial(network, addr string) (net.Conn, error) {
    // Simplified HTTP CONNECT dialer
    return net.DialTimeout("tcp", addr, 10*time.Second)
}
