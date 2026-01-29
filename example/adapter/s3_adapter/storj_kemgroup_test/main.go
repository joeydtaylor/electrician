package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

func main() {
	d := net.Dialer{Timeout: 5 * time.Second}
	conn, err := tls.DialWithDialer(&d, "tcp", "gateway.storjshare.io:443", &tls.Config{
		MinVersion:       tls.VersionTLS13,
		MaxVersion:       tls.VersionTLS13,
		ServerName:       "gateway.storjshare.io",
		CurvePreferences: []tls.CurveID{tls.X25519MLKEM768}, // force it
	})
	if err != nil {
		fmt.Println("MLKEM not supported:", err)
		return
	}
	defer conn.Close()
	fmt.Println("MLKEM supported ✅")
}
