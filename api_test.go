package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	SimpleContentRequest = httptest.NewRequest("GET", "/?offset=0&count=5", nil)
	OffsetContentRequest = httptest.NewRequest("GET", "/?offset=5&count=5", nil)

	errDummy = fmt.Errorf("dummy error for failing provider")
)

type FailingContentProvider struct {
	Source Provider
}

func (FailingContentProvider) GetContent(string, int) ([]*ContentItem, error) {
	return nil, errDummy
}

type LimitedContentProvider struct {
	Provier SampleContentProvider
	limit   int
}

func (l LimitedContentProvider) GetContent(userIP string, limit int) ([]*ContentItem, error) {
	if limit > l.limit {
		limit = l.limit
	}
	return l.Provier.GetContent(userIP, limit)
}
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

func TestFailingResponseFallback(t *testing.T) {
	a := NewApp(map[Provider]Client{
		Provider1: FailingContentProvider{Provider1},
		Provider2: SampleContentProvider{Provider2},
		Provider3: SampleContentProvider{Provider3},
	}, DefaultConfig)
	expectedProviders := []Provider{Provider2, Provider2, Provider2, Provider3}
	content := runRequest(t, a, SimpleContentRequest)
	require.Len(t, content, 4)
	for i := range expectedProviders {
		require.EqualValues(t, expectedProviders[i], content[i].Source)
	}
}

func TestLimitedResponseFallback(t *testing.T) {
	a := NewApp(map[Provider]Client{
		Provider1: LimitedContentProvider{SampleContentProvider{Provider1}, 2},
		Provider2: SampleContentProvider{Provider2},
		Provider3: SampleContentProvider{Provider3},
	}, DefaultConfig)
	expectedProviders := []Provider{Provider1, Provider1, Provider2, Provider3}
	content := runRequest(t, a, SimpleContentRequest)
	require.Len(t, content, 4)
	for i := range expectedProviders {
		require.EqualValues(t, expectedProviders[i], content[i].Source)
	}
}
