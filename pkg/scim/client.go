package scim

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/idf-educ/idm/scim-ctl/pkg/auth"
	"github.com/idf-educ/idm/scim-ctl/pkg/config"
)

// Convert ressource type to SCIM resource name.
func ressourceName(resourceType string) string {
	result := cases.Title(language.English).String(resourceType)
	result = strings.ReplaceAll(result, "-", "")
	if result != "Me" {
		return result + "s"
	}
	return result
}

// Client represents a SCIM client
type Client struct {
	baseURL       string
	httpClient    *http.Client
	accessToken   string
	verbose       bool
	authenticator *auth.Authenticator
	authConfig    *auth.DeviceFlowConfig
}

// Resource represents a generic SCIM resource
type Resource map[string]interface{}

// ListResponse represents a SCIM list response
type ListResponse struct {
	Schemas      []string   `json:"schemas"`
	TotalResults int        `json:"totalResults"`
	StartIndex   int        `json:"startIndex"`
	ItemsPerPage int        `json:"itemsPerPage"`
	Resources    []Resource `json:"Resources"`
}

// ErrorResponse represents a SCIM error response
type ErrorResponse struct {
	Schemas []string `json:"schemas"`
	Detail  string   `json:"detail"`
	Status  int      `json:"status"`
}

// SearchRequest represents a SCIM search request
type SearchRequest struct {
	Schemas      []string `json:"schemas"`
	Filter       string   `json:"filter,omitempty"`
	StartIndex   int      `json:"startIndex,omitempty"`
	Count        int      `json:"count,omitempty"`
	SortBy       string   `json:"sortBy,omitempty"`
	SortOrder    string   `json:"sortOrder,omitempty"`
	Attributes   []string `json:"attributes,omitempty"`
	ExcludedAttr []string `json:"excludedAttributes,omitempty"`
}

// PatchOperation represents a single SCIM patch operation
type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

// NewClient creates a new SCIM client
func NewClient(cfg *config.Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &Client{
		baseURL: strings.TrimSuffix(cfg.Target, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		verbose: cfg.Verbose,
	}, nil
}

// Authenticate performs OAuth 2.0 Device Grant authentication
func (c *Client) Authenticate(ctx context.Context, cfg *config.Config) error {
	authConfig := &auth.DeviceFlowConfig{
		Issuer:       cfg.OIDC.Issuer,
		ClientID:     cfg.OIDC.ClientID,
		ClientSecret: cfg.OIDC.ClientSecret,
		Scopes:       []string{"openid", "profile", "email"},
	}

	authenticator := auth.NewAuthenticator(authConfig)
	accessToken, err := authenticator.GetAccessToken(ctx, c.verbose)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	c.accessToken = accessToken
	c.authenticator = authenticator
	c.authConfig = authConfig
	return nil
}

// GetSchemas retrieves SCIM schemas
func (c *Client) GetSchemas(ctx context.Context) ([]Resource, error) {
	url := c.baseURL + "/Schemas"
	resp, err := c.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var listResp ListResponse
	if err := json.Unmarshal(resp, &listResp); err != nil {
		return nil, fmt.Errorf("failed to decode schemas response: %w", err)
	}

	return listResp.Resources, nil
}

// GetResourceTypes retrieves SCIM resource types
func (c *Client) GetResourceTypes(ctx context.Context) ([]Resource, error) {
	url := c.baseURL + "/ResourceTypes"
	resp, err := c.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var listResp ListResponse
	if err := json.Unmarshal(resp, &listResp); err != nil {
		return nil, fmt.Errorf("failed to decode resource types response: %w", err)
	}

	return listResp.Resources, nil
}

// CreateResource creates a SCIM resource
func (c *Client) CreateResource(ctx context.Context, resourceType string, data Resource) (Resource, error) {
	url := c.baseURL + "/" + ressourceName(resourceType)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resource data: %w", err)
	}

	resp, err := c.doRequest(ctx, "POST", url, jsonData)
	if err != nil {
		return nil, err
	}

	var resource Resource
	if err := json.Unmarshal(resp, &resource); err != nil {
		return nil, fmt.Errorf("failed to decode resource response: %w", err)
	}

	return resource, nil
}

// GetResource retrieves a SCIM resource by ID
func (c *Client) GetResource(ctx context.Context, resourceType, id string, attributes []string) (Resource, error) {
	baseURL := c.baseURL + "/" + ressourceName(resourceType)
	if id != "" {
		baseURL += "/" + id
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add attributes query parameter if specified
	if len(attributes) > 0 {
		query := u.Query()
		query.Set("attributes", strings.Join(attributes, ","))
		u.RawQuery = query.Encode()
	}

	resp, err := c.doRequest(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	var resource Resource
	if err := json.Unmarshal(resp, &resource); err != nil {
		return nil, fmt.Errorf("failed to decode resource response: %w", err)
	}

	return resource, nil
}

// ReplaceResource replaces a SCIM resource
func (c *Client) ReplaceResource(ctx context.Context, resourceType, id string, data Resource) (Resource, error) {
	url := c.baseURL + "/" + ressourceName(resourceType) + "/" + id

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resource data: %w", err)
	}

	resp, err := c.doRequest(ctx, "PUT", url, jsonData)
	if err != nil {
		return nil, err
	}

	var resource Resource
	if err := json.Unmarshal(resp, &resource); err != nil {
		return nil, fmt.Errorf("failed to decode resource response: %w", err)
	}

	return resource, nil
}

// UpdateResource updates a SCIM resource using PATCH
func (c *Client) UpdateResource(ctx context.Context, resourceType, id string, operations []PatchOperation) (Resource, error) {
	url := c.baseURL + "/" + ressourceName(resourceType) + "/" + id

	payload := map[string]interface{}{
		"schemas":    []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
		"Operations": operations,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal patch payload: %w", err)
	}

	resp, err := c.doRequest(ctx, "PATCH", url, jsonData)
	if err != nil {
		return nil, err
	}

	var resource Resource
	if err := json.Unmarshal(resp, &resource); err != nil {
		return nil, fmt.Errorf("failed to decode resource response: %w", err)
	}

	return resource, nil
}

// DeleteResource deletes a SCIM resource
func (c *Client) DeleteResource(ctx context.Context, resourceType, id string) error {
	url := c.baseURL + "/" + ressourceName(resourceType) + "/" + id

	_, err := c.doRequest(ctx, "DELETE", url, nil)
	return err
}

// SearchResources searches SCIM resources
func (c *Client) SearchResources(ctx context.Context, resourceType string, filter string, query string, startIndex, count int, sortBy, sortOrder string, attributes []string) (*ListResponse, error) {
	baseURL := c.baseURL + "/" + ressourceName(resourceType)

	// Use URL parameters for GET request
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	queryParams := u.Query()
	if filter != "" {
		queryParams.Set("filter", filter)
	}
	if query != "" {
		queryParams.Set("q", query)
	}
	if startIndex > 0 {
		queryParams.Set("startIndex", fmt.Sprintf("%d", startIndex))
	}
	if count > 0 {
		queryParams.Set("count", fmt.Sprintf("%d", count))
	}
	if sortBy != "" {
		queryParams.Set("sortBy", sortBy)
	}
	if sortOrder != "" {
		queryParams.Set("sortOrder", sortOrder)
	}
	if len(attributes) > 0 {
		queryParams.Set("attributes", strings.Join(attributes, ","))
	}
	u.RawQuery = queryParams.Encode()

	resp, err := c.doRequest(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	var listResp ListResponse
	if err := json.Unmarshal(resp, &listResp); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	return &listResp, nil
}

// doRequest performs an HTTP request with proper headers and error handling
func (c *Client) doRequest(ctx context.Context, method, url string, body []byte) ([]byte, error) {
	// 1. Ensure we have a valid access token before making the request
	if c.authenticator != nil {
		accessToken, err := c.authenticator.GetAccessTokenSilent(ctx, c.verbose)
		if err != nil {
			if c.verbose {
				fmt.Fprintf(os.Stderr, "Failed to silently renew access token: %v\n", err)
			}
			// We continue with the current token, the call will likely fail with 401
			// which we handle below.
		} else {
			c.accessToken = accessToken
		}
	}

	// 2. Perform the initial request
	respBody, statusCode, status, err := c.performHTTPRequest(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	// 3. Handle 401 Unauthorized by clearing cache and retrying once
	if statusCode == http.StatusUnauthorized && c.authenticator != nil {
		if c.verbose {
			fmt.Fprintf(os.Stderr, "Received 401 Unauthorized, clearing token cache and retrying...\n")
		}

		// Clear cache to remove the invalid token
		c.authenticator.ClearCache(c.verbose)

		// Try to get a new token silently (using refresh token if available)
		accessToken, err := c.authenticator.GetAccessTokenSilent(ctx, c.verbose)
		if err == nil {
			c.accessToken = accessToken
			// Retry the request with the new token
			respBody, statusCode, status, err = c.performHTTPRequest(ctx, method, url, body)
			if err != nil {
				return nil, err
			}
		} else if c.verbose {
			fmt.Fprintf(os.Stderr, "Silent token renewal failed after 401: %v\n", err)
		}
	}

	// 4. Handle error responses
	if statusCode >= 400 {
		var errorResp ErrorResponse
		if err := json.Unmarshal(respBody, &errorResp); err == nil && errorResp.Detail != "" {
			return nil, fmt.Errorf("SCIM error (%d): %s", statusCode, errorResp.Detail)
		}
		return nil, fmt.Errorf("HTTP error %d: %s", statusCode, status)
	}
	return respBody, nil
}

// performHTTPRequest is a helper that carries out the actual HTTP request
func (c *Client) performHTTPRequest(ctx context.Context, method, url string, body []byte) ([]byte, int, string, error) {
	var bodyReader io.Reader = http.NoBody
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, 0, "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/scim+json")
	if method == "POST" || method == "PUT" || method == "PATCH" {
		req.Header.Set("Content-Type", "application/scim+json")
	}
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	if c.verbose {
		fmt.Fprintf(os.Stderr, "→ %s %s\n", method, url)
		if body != nil {
			fmt.Fprintf(os.Stderr, "→ Body: %s\n", string(body))
		}
	}

	if c.verbose {
		dump, err := httputil.DumpRequest(req, true)
		if err == nil {
			fmt.Fprintf(os.Stderr, "→ Request Dump:\n%s\n", dump)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, resp.Status, fmt.Errorf("failed to read response body: %w", err)
	}

	if c.verbose {
		dump, err := httputil.DumpResponse(resp, true)
		if err == nil {
			fmt.Fprintf(os.Stderr, "← %d %s\n", resp.StatusCode, resp.Status)
		} else {
			fmt.Fprintf(os.Stderr, "→ Request Dump:\n%s\n", dump)
		}
	}

	return respBody, resp.StatusCode, resp.Status, nil
}
