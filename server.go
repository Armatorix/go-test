package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

const xForwardedFor = "x-forwarded-for"

var (
	errQueryParamMissing = fmt.Errorf("query parameter missing")
	errMultipleValues    = fmt.Errorf("multiple values")
	errNegativeValue     = fmt.Errorf("negative value")
	errZeroValue         = fmt.Errorf("zero value")
)

// App represents the server's internal state.
// It holds configuration about providers and content.
type App struct {
	ContentClients map[Provider]Client
	Config         ContentMix
	DemandManager  DemandManager
}

// NewApp returns new instance of App with calculated DemandManager.
func NewApp(contentClients map[Provider]Client, config ContentMix) App {
	return App{
		ContentClients: contentClients,
		Config:         config,
		DemandManager:  NewDemandManager(config),
	}
}

// nonNegativeQueryParam parses and validates paramName from url values.
func nonNegativeQueryParam(urlVals url.Values, paramName string) (int, error) {
	paramVals, ok := urlVals[paramName]
	if !ok {
		return 0, fmt.Errorf("param: %s, %w", paramName, errQueryParamMissing)
	}
	if len(paramVals) != 1 {
		return 0, fmt.Errorf("param: %s, quatity: %d, %w", paramName, len(paramVals), errMultipleValues)
	}

	param, err := strconv.ParseInt(paramVals[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("param: %s, %w", paramName, err)
	}
	if param < 0 {
		return 0, fmt.Errorf("param: %s, %w", paramName, errNegativeValue)
	}
	return int(param), nil
}

// requestParams fetches and validates query params of count and offset from request.
func requestParams(req *http.Request) (count, offset int, err error) {
	queryVals := req.URL.Query()
	count, err = nonNegativeQueryParam(queryVals, "count")
	if err != nil {
		return
	}
	if count == 0 {
		err = fmt.Errorf("param: count, %w", errZeroValue)
		return
	}
	offset, err = nonNegativeQueryParam(queryVals, "offset")
	return
}

// getProvidersContent provides content per provider based on calculated demand.
func (a *App) getProvidersContent(count, offset int, userIP string) map[Provider][]*ContentItem {
	var wg sync.WaitGroup
	var m sync.Mutex
	providerDemands := a.DemandManager.ProvidersCounts(count, offset)
	providersContent := make(map[Provider][]*ContentItem)
	wg.Add(len(providerDemands))
	for provider, demand := range providerDemands {
		go func(provider Provider, demand int) {
			content, err := a.ContentClients[provider].GetContent(userIP, demand)
			if err != nil {
				log.Printf("content load failed, provider: %s, err: %v", provider, err)
			} else {
				m.Lock()
				providersContent[provider] = content
				m.Unlock()
			}
			wg.Done()
		}(provider, demand)
	}
	wg.Wait()
	return providersContent
}

// GetContent provides content for user based on app config.
func (a *App) GetContent(count, offset int, userIP string) []*ContentItem {
	providersContent := a.getProvidersContent(count, offset, userIP)
	content := make([]*ContentItem, 0, count)
	for j := 0; j < count; j++ {
		cfg := a.Config[(j+offset)%len(a.Config)]
		switch {
		case len(providersContent[cfg.Type]) > 0:
			content = append(content, providersContent[cfg.Type][0])
			providersContent[cfg.Type] = providersContent[cfg.Type][1:]
		case cfg.Fallback != nil && len(providersContent[*cfg.Fallback]) > 0:
			content = append(content, providersContent[*cfg.Fallback][0])
			providersContent[*cfg.Fallback] = providersContent[*cfg.Fallback][1:]
		default:
			return content
		}
	}
	return content
}

// userIP returns user IP based on xff header.
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For
func userIP(req *http.Request) string {
	return strings.Split(req.Header.Get(xForwardedFor), ",")[0]
}

func (a App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s", req.Method, req.URL.String())
	count, offset, err := requestParams(req)
	if err != nil || count == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	items := a.GetContent(count, offset, userIP(req))
	if err := json.NewEncoder(w).Encode(items); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
