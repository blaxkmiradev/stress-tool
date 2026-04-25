// pkg/client/http_client.go
package client

import (
    "crypto/tls"
    "net"
    "net/http"
    "net/url"
    "time"

    "github.com/stresstest/layer7-flood/pkg/config"
)

type HttpClient struct {
    client  *http.Client
    config  *config.Config
}

func NewHttpClient(cfg *config.Config) *HttpClient {
    transport := &http.Transport{
        DialContext: (&net.Dialer{
            Timeout:   30 * time.Second,
            KeepAlive: 30 * time.Second,
        }).DialContext,
        MaxIdleConns:          1000,
        MaxIdleConnsPerHost:   1000,
        IdleConnTimeout:       90 * time.Second,
        TLSHandshakeTimeout:   10 * time.Second,
        ExpectContinueTimeout: 1 * time.Second,
        DisableKeepAlives:     !cfg.Client.KeepAlive,
        DisableCompression:    false,
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: true,
            Renegotiation:      tls.RenegotiateFreelyAsClient,
        },
    }
    
    if cfg.Proxy.Enabled {
        proxyURL, _ := url.Parse(cfg.Proxy.URL)
        transport.Proxy = http.ProxyURL(proxyURL)
    }
    
    return &HttpClient{
        client: &http.Client{
            Transport: transport,
            Timeout:   time.Duration(cfg.Client.Timeout) * time.Second,
            CheckRedirect: func(req *http.Request, via []*http.Request) error {
                if len(via) >= 10 {
                    return http.ErrUseLastResponse
                }
                return nil
            },
        },
        config: cfg,
    }
}

func (h *HttpClient) Do(req *http.Request) (*http.Response, error) {
    return h.client.Do(req)
}

func (h *HttpClient) Close() {
    h.client.CloseIdleConnections()
}
