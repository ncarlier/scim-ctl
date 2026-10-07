package scim

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ncarlier/scim-ctl/pkg/config"
)

func TestClientWithoutAuth(t *testing.T) {
	// Create mock SCIM server
	var receivedAuthHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuthHeader = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/scim+json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"schemas":["urn:ietf:params:scim:api:messages:2.0:ListResponse"],"totalResults":0,"itemsPerPage":0,"startIndex":1,"Resources":[]}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		Target:  server.URL,
		Timeout: 5,
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	if client.HasAuth() {
		t.Errorf("client.HasAuth() = true, want false")
	}

	ctx := context.Background()
	if err := client.Authenticate(ctx, cfg); err != nil {
		t.Fatalf("client.Authenticate() failed: %v", err)
	}

	if client.HasAuth() {
		t.Errorf("client.HasAuth() after Authenticate() = true, want false")
	}

	schemas, err := client.GetSchemas(ctx)
	if err != nil {
		t.Fatalf("GetSchemas() failed: %v", err)
	}

	if len(schemas) != 0 {
		t.Errorf("len(schemas) = %d, want 0", len(schemas))
	}

	if receivedAuthHeader != "" {
		t.Errorf("received Authorization header = %q, want empty", receivedAuthHeader)
	}
}

func TestClientWithoutAuthWithExtraHeader(t *testing.T) {
	var receivedAuthHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuthHeader = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/scim+json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"schemas":["urn:ietf:params:scim:api:messages:2.0:ListResponse"],"totalResults":0,"itemsPerPage":0,"startIndex":1,"Resources":[]}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		Target:  server.URL,
		Timeout: 5,
		ExtraHeaders: map[string]string{
			"Authorization": "Bearer static-token",
		},
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	if err := client.Authenticate(ctx, cfg); err != nil {
		t.Fatalf("client.Authenticate() failed: %v", err)
	}

	_, err = client.GetSchemas(ctx)
	if err != nil {
		t.Fatalf("GetSchemas() failed: %v", err)
	}

	if receivedAuthHeader != "Bearer static-token" {
		t.Errorf("received Authorization header = %q, want 'Bearer static-token'", receivedAuthHeader)
	}
}
