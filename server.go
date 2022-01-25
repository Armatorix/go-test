package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

// App represents the server's internal state.
// It holds configuration about providers and content.

var (
	errQueryParamMissing = fmt.Errorf("query parameter missing")
	errMultipleValues    = fmt.Errorf("multiple values")
	errNegativeValue     = fmt.Errorf("negative value")
)

type App struct {
	ContentClients map[Provider]Client
	Config         ContentMix
}

func nonNegativeQueryParam(vals url.Values, paramName string) (int64, error) {
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
	return param, nil
}

func requestParams(req *http.Request) (count int64, offset int64, err error) {
	queryVals := req.URL.Query()
	count, err = nonNegativeQueryParam(queryVals, "count")
	if err != nil {
		return
	}
	offset, err = nonNegativeQueryParam(queryVals, "offset")
	return
}

func (a App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s", req.Method, req.URL.String())
	count, offset, err := requestParams(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	log.Printf("count %d offset %d", count, offset)
	w.WriteHeader(http.StatusOK)
}
