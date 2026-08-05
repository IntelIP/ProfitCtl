package exa

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientSearch_SendsOfficialDomainRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodPost, request.Method)
		assert.Equal(t, "test-key", request.Header.Get("x-api-key"))

		var search SearchRequest
		require.NoError(t, json.NewDecoder(request.Body).Decode(&search))
		assert.Equal(t, "OpenRouter pricing", search.Query)
		assert.Equal(t, []string{"openrouter.ai"}, search.IncludeDomains)
		assert.True(t, search.Contents.Highlights)

		writer.Header().Set("Content-Type", "application/json")
		_, err := writer.Write([]byte(`{"requestId":"exa-request","results":[{"title":"Terra","url":"https://openrouter.ai/openai/gpt-5.6-terra","highlights":["$1 / $6 per 1M"]}]}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	client := &Client{apiKey: "test-key", endpoint: server.URL, httpClient: server.Client()}
	response, err := client.Search(context.Background(), SearchRequest{
		Query:          "OpenRouter pricing",
		NumResults:     3,
		IncludeDomains: []string{"openrouter.ai"},
	})

	require.NoError(t, err)
	assert.Equal(t, "exa-request", response.RequestID)
	require.Len(t, response.Results, 1)
	assert.Equal(t, "https://openrouter.ai/openai/gpt-5.6-terra", response.Results[0].URL)
}

func TestNewClient_UsesBoundedRequestTimeout(t *testing.T) {
	client, err := NewClient(" test-key ")

	require.NoError(t, err)
	assert.Equal(t, "test-key", client.apiKey)
	require.NotNil(t, client.httpClient)
	assert.Equal(t, defaultRequestTimeout, client.httpClient.Timeout)
}
