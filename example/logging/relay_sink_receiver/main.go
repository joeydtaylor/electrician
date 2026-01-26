package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
	"github.com/joeydtaylor/electrician/pkg/logschema"
)

const relayBufferSize = 4096

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envOrInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envOrDuration(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func envOrBool(k string, def bool) bool {
	if v := os.Getenv(k); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	runFor := envOrDuration("LOG_RELAY_DURATION", 0)
	if runFor > 0 {
		timer := time.NewTimer(runFor)
		defer timer.Stop()
		go func() {
			select {
			case <-timer.C:
				stop()
			case <-ctx.Done():
			}
		}()
	}

	logger := builder.NewLogger(builder.LoggerWithDevelopment(true))
	workers := runtime.GOMAXPROCS(0)

	wire := builder.NewWire(
		ctx,
		builder.WireWithLogger[logschema.LogRecord](logger),
		builder.WireWithConcurrencyControl[logschema.LogRecord](relayBufferSize, workers),
	)
	if err := wire.Start(ctx); err != nil {
		log.Fatalf("wire start: %v", err)
	}

	tlsCfg := builder.NewTlsServerConfig(
		true,
		envOr("TLS_CERT", "../tls/server.crt"),
		envOr("TLS_KEY", "../tls/server.key"),
		envOr("TLS_CA", "../tls/ca.crt"),
		"localhost",
		tls.VersionTLS12, tls.VersionTLS13,
	)

	recv := builder.NewReceivingRelay(
		ctx,
		builder.ReceivingRelayWithAddress[logschema.LogRecord](envOr("RX_ADDR", "localhost:50090")),
		builder.ReceivingRelayWithBufferSize[logschema.LogRecord](relayBufferSize),
		builder.ReceivingRelayWithLogger[logschema.LogRecord](logger),
		builder.ReceivingRelayWithOutput(wire),
		builder.ReceivingRelayWithTLSConfig[logschema.LogRecord](tlsCfg),
	)

	if err := recv.Start(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Log relay receiver up. Ctrl+C to stop.")
	<-ctx.Done()
	fmt.Println("Shutting down...")

	recv.Stop()

	fmt.Println("---- Log Relay ----")
	if b, err := wire.LoadAsJSONArray(); err == nil {
		var records []logschema.LogRecord
		if err := json.Unmarshal(b, &records); err != nil {
			fmt.Println("err:", err)
			fmt.Println(string(b))
		} else {
			fmt.Printf("Captured %d log entries\n", len(records))

			sample := envOrInt("LOG_RELAY_SAMPLE", 5)
			if sample > len(records) {
				sample = len(records)
			}
			for i := 0; i < sample; i++ {
				line, err := json.Marshal(records[i])
				if err != nil {
					fmt.Println("err:", err)
					continue
				}
				fmt.Printf("log[%d]: %s\n", i+1, string(line))
			}

			if envOrBool("LOG_RELAY_DUMP", false) {
				fmt.Println(string(b))
			}

			if outPath := envOr("LOG_RELAY_OUTPUT", ""); outPath != "" {
				if err := writeJSONLines(outPath, records); err != nil {
					fmt.Println("err:", err)
				} else {
					fmt.Printf("Wrote %d log entries to %s\n", len(records), outPath)
				}
			}
		}
	} else {
		fmt.Println("err:", err)
	}

	fmt.Println("Done.")
}

func writeJSONLines(path string, records []logschema.LogRecord) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, record := range records {
		line, err := json.Marshal(record)
		if err != nil {
			return err
		}
		if _, err := writer.Write(line); err != nil {
			return err
		}
		if _, err := writer.WriteString("\n"); err != nil {
			return err
		}
	}
	return writer.Flush()
}
