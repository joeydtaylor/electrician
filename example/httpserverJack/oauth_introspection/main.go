package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

type MyRequest struct {
	Name string `json:"name"`
}

type MyResponse struct {
	Greeting string `json:"greeting"`
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverAddr := envOr("HTTP_ADDR", ":8443")
	introspectAddr := envOr("INTROSPECT_ADDR", "localhost:8081")
	clientID := envOr("INTROSPECT_CLIENT_ID", "example-client")
	clientSecret := envOr("INTROSPECT_CLIENT_SECRET", "example-secret")
	requiredScope := envOr("INTROSPECT_SCOPE", "write:hello")
	expectedToken := envOr("INTROSPECT_TOKEN", "token-123")

	logger := builder.NewLogger(
		builder.LoggerWithLevel("info"),
		builder.LoggerWithDevelopment(true),
	)

	introspectSrv := &http.Server{
		Addr: introspectAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			user, pass, ok := r.BasicAuth()
			if !ok || user != clientID || pass != clientSecret {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if err := r.ParseForm(); err != nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}

			token := strings.TrimSpace(r.FormValue("token"))
			active := token == expectedToken
			scope := ""
			if active {
				scope = requiredScope
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"active": active,
				"scope":  scope,
			})
		}),
	}

	go func() {
		if err := introspectSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("introspection server error: %v", err)
			stop()
		}
	}()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = introspectSrv.Shutdown(shutdownCtx)
	}()

	introspectOpts := builder.NewHTTPServerOAuth2IntrospectionOptions(
		"http://"+introspectAddr+"/introspect",
		"basic",
		clientID,
		clientSecret,
		"",
		30,
	)
	introspectOpts.RequiredScopes = []string{requiredScope}
	authOpts := builder.NewHTTPServerAuthenticationOptionsOAuth2(introspectOpts)

	tlsConfig := builder.NewTlsServerConfig(
		true,
		envOr("TLS_CERT", "../tls/server.crt"),
		envOr("TLS_KEY", "../tls/server.key"),
		envOr("TLS_CA", "../tls/ca.crt"),
		"localhost",
		tls.VersionTLS12, tls.VersionTLS13,
	)

	server := builder.NewHTTPServer(ctx,
		builder.HTTPServerWithAddress[MyRequest](serverAddr),
		builder.HTTPServerWithServerConfig[MyRequest](http.MethodPost, "/hello"),
		builder.HTTPServerWithLogger[MyRequest](logger),
		builder.HTTPServerWithTimeout[MyRequest](5*time.Second),
		builder.HTTPServerWithTLS[MyRequest](*tlsConfig),
		builder.HTTPServerWithAuthenticationOptions[MyRequest](authOpts),
		builder.HTTPServerWithAuthRequired[MyRequest](true),
	)

	handleRequest := func(ctx context.Context, req MyRequest) (builder.HTTPServerResponse, error) {
		responseData := MyResponse{Greeting: "Hello, " + req.Name}
		bodyBytes, err := json.Marshal(responseData)
		if err != nil {
			return builder.HTTPServerResponse{}, fmt.Errorf("failed to marshal response: %w", err)
		}

		return builder.HTTPServerResponse{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: bodyBytes,
		}, nil
	}

	fmt.Printf("[Main] OAuth2 introspection: http://%s/introspect\n", introspectAddr)
	fmt.Printf("[Main] HTTPS server: https://localhost%s/hello\n", serverAddr)
	fmt.Printf("[Main] curl: curl -k -X POST https://localhost%s/hello -H 'Authorization: Bearer %s' -d '{\"name\":\"Joey\"}'\n", serverAddr, expectedToken)

	if err := server.Serve(ctx, handleRequest); err != nil {
		fmt.Printf("[Main] Server stopped with error: %v\n", err)
	} else {
		fmt.Println("[Main] Server stopped gracefully.")
	}
}
