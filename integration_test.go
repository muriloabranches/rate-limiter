package main

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

type fixedIPRoundTripper struct {
	Proxied http.RoundTripper
	IP      string
}

func (f *fixedIPRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.RemoteAddr = net.JoinHostPort(f.IP, "12345")
	return f.Proxied.RoundTrip(req)
}

func setupTestServer() *httptest.Server {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, // Use a different DB for testing
	})
	client.FlushDB(context.Background())

	persistence := NewRedisPersistence(client)
	rateLimiter := NewRateLimiter(persistence)

	handler := RateLimiterMiddleware(rateLimiter, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	}))

	return httptest.NewServer(handler)
}

func setupEnv() {
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("REDIS_PASSWORD", "")
	os.Setenv("REDIS_DB", "1")
	os.Setenv("IP_RATE_LIMIT", "3")
	os.Setenv("TOKEN_RATE_LIMIT", "5")
	os.Setenv("BLOCK_DURATION", "30")   // in seconds
	os.Setenv("RATE_LIMIT_WINDOW", "3") // in seconds
}

func TestRateLimiterIntegration_BlockByIP(t *testing.T) {
	setupEnv()
	server := setupTestServer()
	defer server.Close()

	client := &http.Client{
		Transport: &fixedIPRoundTripper{
			Proxied: http.DefaultTransport,
			IP:      "192.168.1.1",
		},
	}

	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest("GET", server.URL, nil)
		resp, err := client.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
}

func TestRateLimiterIntegration_BlockByToken(t *testing.T) {
	setupEnv()
	server := setupTestServer()
	defer server.Close()

	client := &http.Client{}
	token := "test-token"

	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest("GET", server.URL, nil)
		req.Header.Set("API_KEY", token)
		resp, err := client.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set("API_KEY", token)
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
}
