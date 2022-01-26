package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

// App represents the server's internal state.
// It holds configuration about providers and content.

var (
	errQueryParamMissing = fmt.Errorf("query parameter missing")
	errMultipleValues    = fmt.Errorf("multiple values")
	errNegativeValue     = fmt.Errorf("negative value")
	errZeroValue         = fmt.Errorf("zero value")
)

type App struct {
	ContentClients map[Provider]Client
	Config         ContentMix
	ContentDemand  ContentDemand
}

func NewApp(contentClients map[Provider]Client, config ContentMix) App {
	return App{
		ContentClients: contentClients,
		Config:         config,
		ContentDemand:  NewContentDemand(config),
	}
}

func nonNegativeQueryParam(vals url.Values, paramName string) (int, error) {
	paramVals, ok := vals[paramName]
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

// TODO: refacotr : split content batch with pointer to elements.
func (a *App) GetContent(count, offset int, userIP string) []*ContentItem {
	var wg sync.WaitGroup
	var rwm sync.Mutex
	providerDemands := a.ContentDemand.ProvidersCounts(count, offset)
	providersContent := make(map[Provider][]*ContentItem)
	wg.Add(len(providerDemands))
	for provider, demand := range providerDemands {
		go func(provider Provider, demand int) {
			content, err := a.ContentClients[provider].GetContent(userIP, demand)
			if err != nil {
				log.Printf("content load failed, provider: %s, err: %v", provider, err)
			} else {
				rwm.Lock()
				providersContent[provider] = content
				rwm.Unlock()
			}
			wg.Done()
		}(provider, demand)
	}
	wg.Wait()

	content := make([]*ContentItem, 0, count)
	for j := 0; j < count; j++ {
		cfg := a.Config[(j+offset)%len(a.Config)]
		if len(providersContent[cfg.Type]) > 0 {
			content = append(content, providersContent[cfg.Type][0])
			providersContent[cfg.Type] = providersContent[cfg.Type][1:]
		} else if len(providersContent[*cfg.Fallback]) > 0 {
			content = append(content, providersContent[*cfg.Fallback][0])
			providersContent[*cfg.Fallback] = providersContent[*cfg.Fallback][1:]
		} else {
			break
		}
	}
	return content
}

func (a App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s", req.Method, req.URL.String())
	count, offset, err := requestParams(req)
	if err != nil || count == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// TODO: fetch IP
	items := a.GetContent(count, offset, "testip")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
