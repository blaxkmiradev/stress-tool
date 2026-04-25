// pkg/utils/random.go
package utils

import (
    "crypto/rand"
    "fmt"
    "math/big"
    "time"
)

var (
    letters    = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
    userAgents = []string{
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
        "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
        "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15",
        "Mozilla/5.0 (Windows NT 10.0; rv:91.0) Gecko/20100101 Firefox/91.0",
    }
)

func RandomString(n int) string {
    b := make([]rune, n)
    for i := range b {
        num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
        b[i] = letters[num.Int64()]
    }
    return string(b)
}

func RandomInt(min, max int) int {
    num, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
    return min + int(num.Int64())
}

func RandomIP() string {
    return fmt.Sprintf("%d.%d.%d.%d",
        RandomInt(1, 255),
        RandomInt(0, 255),
        RandomInt(0, 255),
        RandomInt(1, 254),
    )
}

func RandomUserAgent() string {
    return userAgents[RandomInt(0, len(userAgents)-1)]
}

func RandomPort() int {
    return RandomInt(1024, 65535)
}
