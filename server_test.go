package main

import (
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequestParamsParsing(t *testing.T) {
	cases := []struct {
		reqTarget string
		count     int64
		offset    int64
	}{
		{
			reqTarget: "/?offset=0&count=0",
			offset:    0,
			count:     0,
		},
		{
			reqTarget: "/?offset=21&count=73",
			offset:    21,
			count:     73,
		},
	}

	for _, c := range cases {
		t.Run(c.reqTarget, func(t *testing.T) {
			count, offset, err := requestParams(httptest.NewRequest("GET", c.reqTarget, nil))
			require.Equal(t, c.count, count)
			require.Equal(t, c.offset, offset)
			require.NoError(t, err)
		})
	}
}

func TestFailingRequestParamsParsing(t *testing.T) {
	cases := []struct {
		reqTarget string
		err       error
	}{
		{
			reqTarget: "/",
			err:       errQueryParamMissing,
		},
		{
			reqTarget: "/?offset=21",
			err:       errQueryParamMissing,
		},
		{
			reqTarget: "/?count=73",
			err:       errQueryParamMissing,
		},
		{
			reqTarget: "/?count=73&count=21",
			err:       errMultipleValues,
		},
		{
			reqTarget: "/?offset=-21&count=0",
			err:       errNegativeValue,
		},
		{
			reqTarget: "/?offset=0&count=-21",
			err:       errNegativeValue,
		},
		{
			reqTarget: "/?count=dummy",
			err:       strconv.ErrSyntax,
		},
	}

	for _, c := range cases {
		t.Run(c.reqTarget, func(t *testing.T) {
			_, _, err := requestParams(httptest.NewRequest("GET", c.reqTarget, nil))
			require.ErrorIs(t, err, c.err)
		})
	}
}
