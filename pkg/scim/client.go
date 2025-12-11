package scim

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/idf-educ/idm/scim-ctl/pkg/auth"
	"github.com/idf-educ/idm/scim-ctl/pkg/config"
)

// titleCase converts a string to title case using proper Unicode handling
func titleCase(s string) string {
	caser := cases.Title(language.English)
	return caser.String(s)
}

// Client represents a SCIM client
type Client struct {
	baseURL     string
	httpClient  *http.Client
	accessToken string
	verbose     bool
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

// CreateResource creates a SCIM resource
func (c *Client) CreateResource(ctx context.Context, resourceType string, data Resource) (Resource, error) {
	url := c.baseURL + "/" + titleCase(resourceType) + "s"

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
	baseURL := c.baseURL + "/" + titleCase(resourceType) + "s/" + id

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

// UpdateResource updates a SCIM resource
func (c *Client) UpdateResource(ctx context.Context, resourceType, id string, data Resource) (Resource, error) {
	url := c.baseURL + "/" + titleCase(resourceType) + "s/" + id

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

// DeleteResource deletes a SCIM resource
func (c *Client) DeleteResource(ctx context.Context, resourceType, id string) error {
	url := c.baseURL + "/" + titleCase(resourceType) + "s/" + id

	_, err := c.doRequest(ctx, "DELETE", url, nil)
	return err
}

// SearchResources searches SCIM resources
func (c *Client) SearchResources(ctx context.Context, resourceType string, filter string, startIndex, count int, sortBy, sortOrder string, attributes []string) (*ListResponse, error) {
	baseURL := c.baseURL + "/" + titleCase(resourceType) + "s"

	// Use URL parameters for GET request
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	query := u.Query()
	if filter != "" {
		query.Set("filter", filter)
	}
	if startIndex > 0 {
		query.Set("startIndex", fmt.Sprintf("%d", startIndex))
	}
	if count > 0 {
		query.Set("count", fmt.Sprintf("%d", count))
	}
	if sortBy != "" {
		query.Set("sortBy", sortBy)
	}
	if sortOrder != "" {
		query.Set("sortOrder", sortOrder)
	}
	if len(attributes) > 0 {
		query.Set("attributes", strings.Join(attributes, ","))
	}
	u.RawQuery = query.Encode()

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
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/scim+json")
	req.Header.Set("Content-Type", "application/scim+json")

	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	if c.verbose {
		fmt.Printf("→ %s %s\n", method, url)
		if body != nil {
			fmt.Printf("→ Body: %s\n", string(body))
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if c.verbose {
		fmt.Printf("← %d %s\n", resp.StatusCode, resp.Status)
		if len(respBody) > 0 {
			fmt.Printf("← Body: %s\n", string(respBody))
		}
	}

	// Handle error responses
	if resp.StatusCode >= 400 {
		var errorResp ErrorResponse
		if err := json.Unmarshal(respBody, &errorResp); err == nil && errorResp.Detail != "" {
			return nil, fmt.Errorf("SCIM error (%d): %s", resp.StatusCode, errorResp.Detail)
		}
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, resp.Status)
	}

	return respBody, nil
}
