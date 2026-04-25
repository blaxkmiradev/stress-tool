// pkg/config/config.go
package config

import (
    "io/ioutil"
    "strings"

    "gopkg.in/yaml.v3"
)

type Config struct {
    Target   TargetConfig   `yaml:"target"`
    Attack   AttackConfig   `yaml:"attack"`
    Client   ClientConfig   `yaml:"client"`
    Proxy    ProxyConfig    `yaml:"proxy"`
    Headers  HeadersConfig  `yaml:"headers"`
    Bypass   BypassConfig   `yaml:"bypass"`
}

type TargetConfig struct {
    URL      string `yaml:"url"`
    Host     string `yaml:"host"`
    Hostname string `yaml:"hostname"`
    Port     int    `yaml:"port"`
    SSL      bool   `yaml:"ssl"`
}

type AttackConfig struct {
    Threads       int `yaml:"threads"`
    Duration      int `yaml:"duration"`
    RatePerThread int `yaml:"rate_per_thread"`
}

type ClientConfig struct {
    Timeout       int  `yaml:"timeout"`
    KeepAlive     bool `yaml:"keep_alive"`
    MaxRetries    int  `yaml:"max_retries"`
    RetryDelay    int  `yaml:"retry_delay"`
}

type ProxyConfig struct {
    Enabled bool   `yaml:"enabled"`
    URL     string `yaml:"url"`
    Type    string `yaml:"type"`
    Pool    []string `yaml:"pool"`
    Rotate  bool   `yaml:"rotate"`
}

type HeadersConfig struct {
    RandomUserAgents   bool     `yaml:"random_user_agents"`
    RandomHeaders      bool     `yaml:"random_headers"`
    CustomHeaders      map[string]string `yaml:"custom_headers"`
    RefererSpam        bool     `yaml:"referer_spam"`
}

type BypassConfig struct {
    UseHTTP2          bool `yaml:"use_http2"`
    FragmentPackets   bool `yaml:"fragment_packets"`
    PipelineRequests  int  `yaml:"pipeline_requests"`
}

func Load(path string) (*Config, error) {
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    var cfg Config
    err = yaml.Unmarshal(data, &cfg)
    if err != nil {
        return nil, err
    }
    
    // Parse target URL components
    if cfg.Target.URL != "" {
        if strings.HasPrefix(cfg.Target.URL, "https://") {
            cfg.Target.SSL = true
            cfg.Target.Port = 443
        } else if strings.HasPrefix(cfg.Target.URL, "http://") {
            cfg.Target.SSL = false
            cfg.Target.Port = 80
        }
        
        // Extract hostname
        urlWithoutProto := strings.TrimPrefix(cfg.Target.URL, "https://")
        urlWithoutProto = strings.TrimPrefix(urlWithoutProto, "http://")
        parts := strings.Split(urlWithoutProto, "/")
        cfg.Target.Hostname = parts[0]
        cfg.Target.Host = cfg.Target.Hostname
        
        if strings.Contains(cfg.Target.Hostname, ":") {
            hostParts := strings.Split(cfg.Target.Hostname, ":")
            cfg.Target.Hostname = hostParts[0]
            cfg.Target.Port, _ = strconv.Atoi(hostParts[1])
        }
    }
    
    return &cfg, nil
}
