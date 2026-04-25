// pkg/client/transport.go
package client

import (
    "bufio"
    "bytes"
    "net"
    "net/http"
    "time"
)

type RawTransport struct {
    conn    net.Conn
    timeout time.Duration
}

func NewRawTransport(addr string, timeout time.Duration) (*RawTransport, error) {
    conn, err := net.DialTimeout("tcp", addr, timeout)
    if err != nil {
        return nil, err
    }
    return &RawTransport{conn: conn, timeout: timeout}, nil
}

func (r *RawTransport) SendRawRequest(request []byte) ([]byte, error) {
    r.conn.SetWriteDeadline(time.Now().Add(r.timeout))
    _, err := r.conn.Write(request)
    if err != nil {
        return nil, err
    }
    
    r.conn.SetReadDeadline(time.Now().Add(r.timeout))
    reader := bufio.NewReader(r.conn)
    return reader.ReadBytes('\n')
}

func (r *RawTransport) SendPartialRequest(partial []byte) error {
    r.conn.SetWriteDeadline(time.Now().Add(r.timeout))
    _, err := r.conn.Write(partial)
    return err
}

func (r *RawTransport) Close() {
    if r.conn != nil {
        r.conn.Close()
    }
}

func SendHTTPRequestRaw(method, host, path string, headers map[string]string, body []byte) []byte {
    var buffer bytes.Buffer
    
    buffer.WriteString(method + " " + path + " HTTP/1.1\r\n")
    buffer.WriteString("Host: " + host + "\r\n")
    
    for k, v := range headers {
        buffer.WriteString(k + ": " + v + "\r\n")
    }
    
    if len(body) > 0 {
        buffer.WriteString("Content-Length: " + string(rune(len(body))) + "\r\n")
        buffer.WriteString("\r\n")
        buffer.Write(body)
    } else {
        buffer.WriteString("\r\n")
    }
    
    return buffer.Bytes()
}
