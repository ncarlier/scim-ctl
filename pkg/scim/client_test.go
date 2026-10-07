package scim

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestSearchResourcesCountParam(t *testing.T) {
	tests := []struct {
		name          string
		count         *int
		wantHasCount  bool
		wantCountVal  string
	}{
		{
			name:         "count nil (not provided)",
			count:        nil,
			wantHasCount: false,
		},
		{
			name:         "count 0 (count only)",
			count:        intPtr(0),
			wantHasCount: true,
			wantCountVal: "0",
		},
		{
			name:         "count 10",
			count:        intPtr(10),
			wantHasCount: true,
			wantCountVal: "10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedURL string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedURL = r.URL.String()
				w.Header().Set("Content-Type", "application/scim+json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"schemas":["urn:ietf:params:scim:api:messages:2.0:ListResponse"],"totalResults":42,"itemsPerPage":0,"startIndex":1,"Resources":[]}`))
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

			_, err = client.SearchResources(context.Background(), "Users", "", "", 0, tt.count, "", "", nil, nil)
			if err != nil {
				t.Fatalf("SearchResources() failed: %v", err)
			}

			u, err := url.Parse(receivedURL)
			if err != nil {
				t.Fatalf("failed to parse received URL: %v", err)
			}

			hasCount := u.Query().Has("count")
			if hasCount != tt.wantHasCount {
				t.Errorf("has count query param = %v, want %v", hasCount, tt.wantHasCount)
			}
			if tt.wantHasCount {
				gotVal := u.Query().Get("count")
				if gotVal != tt.wantCountVal {
					t.Errorf("count query param = %q, want %q", gotVal, tt.wantCountVal)
				}
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}
