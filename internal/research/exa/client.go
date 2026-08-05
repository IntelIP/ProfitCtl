// Package exa provides the narrow official-domain pricing lookup used by profitctl assess.
package exa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultSearchURL = "https://api.exa.ai/search"
const defaultRequestTimeout = 30 * time.Second

// SearchRequest contains the small supported subset of Exa Search used for pricing receipts.
type SearchRequest struct {
	Query          string   `json:"query"`
	Type           string   `json:"type"`
	NumResults     int      `json:"numResults"`
	IncludeDomains []string `json:"includeDomains"`
	Contents       Contents `json:"contents"`
}

// Contents requests concise source excerpts rather than full crawled pages.
type Contents struct {
	Highlights bool `json:"highlights"`
}

// SearchResult is the source information needed to preserve provider-price provenance.
type SearchResult struct {
	Title      string   `json:"title"`
	URL        string   `json:"url"`
	Highlights []string `json:"highlights"`
}

// SearchResponse is the relevant Exa Search response shape.
type SearchResponse struct {
	RequestID string         `json:"requestId"`
	Results   []SearchResult `json:"results"`
}

// Searcher supports a stubbed, no-network command test.
type Searcher interface {
	Search(context.Context, SearchRequest) (SearchResponse, error)
}

// Client calls the Exa Search API.
type Client struct {
	apiKey     string
	endpoint   string
	httpClient *http.Client
}

// NewClient creates a client for a caller-provided key. The key is never persisted.
func NewClient(apiKey string) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("EXA_API_KEY is required")
	}
	return &Client{
		apiKey:     strings.TrimSpace(apiKey),
		endpoint:   defaultSearchURL,
		httpClient: &http.Client{Timeout: defaultRequestTimeout},
	}, nil
}

// Search retrieves official-domain pricing excerpts through Exa.
func (c *Client) Search(ctx context.Context, search SearchRequest) (SearchResponse, error) {
	if strings.TrimSpace(search.Query) == "" {
		return SearchResponse{}, fmt.Errorf("search query is required")
	}
	if len(search.IncludeDomains) == 0 {
		return SearchResponse{}, fmt.Errorf("at least one official domain is required")
	}
	if search.NumResults < 1 {
		search.NumResults = 1
	}
	if search.Type == "" {
		search.Type = "auto"
	}
	search.Contents.Highlights = true

	body, err := json.Marshal(search)
	if err != nil {
		return SearchResponse{}, fmt.Errorf("marshal Exa search request: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return SearchResponse{}, fmt.Errorf("create Exa search request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("x-api-key", c.apiKey)

	client := c.httpClient
	if client == nil {
		client = http.DefaultClient
	}
	response, err := client.Do(request)
	if err != nil {
		return SearchResponse{}, fmt.Errorf("call Exa search: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return SearchResponse{}, fmt.Errorf("Exa search returned %s: %s", response.Status, strings.TrimSpace(string(message)))
	}

	var result SearchResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return SearchResponse{}, fmt.Errorf("decode Exa search response: %w", err)
	}
	return result, nil
}
