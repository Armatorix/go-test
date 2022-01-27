package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	SimpleContentRequest = httptest.NewRequest("GET", "/?offset=0&count=5", nil)
	OffsetContentRequest = httptest.NewRequest("GET", "/?offset=5&count=5", nil)
)

func runRequest(t *testing.T, srv http.Handler, r *http.Request) (content []*ContentItem) {
	response := httptest.NewRecorder()
	srv.ServeHTTP(response, r)

	require.Equal(t, http.StatusOK, response.Code)

	decoder := json.NewDecoder(response.Body)
	err := decoder.Decode(&content)
	require.NoError(t, err)

	return content
}

func TestBadParam(t *testing.T) {
	response := httptest.NewRecorder()
	app.ServeHTTP(response, httptest.NewRequest("GET", "/?offset=-1&count=0", nil))

	require.Equal(t, http.StatusBadRequest, response.Code)
}

func TestResponseCount(t *testing.T) {
	content := runRequest(t, app, SimpleContentRequest)
	require.Len(t, content, 5)
}
func TestResponseOrder(t *testing.T) {
	content := runRequest(t, app, SimpleContentRequest)
	require.Len(t, content, 5)

	for i, item := range content {
		require.EqualValues(t, DefaultConfig[i%len(DefaultConfig)].Type, item.Source)
	}
}

func TestOffsetResponseOrder(t *testing.T) {
	content := runRequest(t, app, OffsetContentRequest)
	require.Len(t, content, 5)

	for j, item := range content {
		i := j + 5
		require.EqualValues(t, DefaultConfig[i%len(DefaultConfig)].Type, item.Source)
	}
}
