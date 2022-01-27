package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testProviders    = []Provider{Provider("t1"), Provider("t2"), Provider("t3")}
	noFallbackConfig = []ContentConfig{
		{Type: testProviders[0]},
		{Type: testProviders[0]},
		{Type: testProviders[1]},
		{Type: testProviders[2]},
		{Type: testProviders[1]},
	}
)

func TestNewDemandManager(t *testing.T) {
	cases := []struct {
		name                string
		config              ContentMix
		expectedLookupTable []map[Provider]int
	}{
		{
			name:   "default config",
			config: DefaultConfig,
			expectedLookupTable: []map[Provider]int{
				{Provider1: 1, Provider2: 1},
				{Provider1: 2, Provider2: 2},
				{Provider1: 2, Provider2: 3, Provider3: 1},
				{Provider1: 3, Provider2: 3, Provider3: 2},
				{Provider1: 4, Provider2: 3, Provider3: 2},
				{Provider1: 5, Provider2: 4, Provider3: 2},
				{Provider1: 6, Provider2: 5, Provider3: 2},
				{Provider1: 6, Provider2: 6, Provider3: 3},
			},
		},
		{
			name:   "no fallback config",
			config: noFallbackConfig,
			expectedLookupTable: []map[Provider]int{
				{testProviders[0]: 1},
				{testProviders[0]: 2},
				{testProviders[0]: 2, testProviders[1]: 1},
				{testProviders[0]: 2, testProviders[1]: 1, testProviders[2]: 1},
				{testProviders[0]: 2, testProviders[1]: 2, testProviders[2]: 1},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dm := NewDemandManager(c.config)
			require.EqualValues(t, c.expectedLookupTable, dm.demandLookup)
		})
	}
}
func TestDefaultConfigDemandManagerCalculation(t *testing.T) {
	cases := []struct {
		offset         int
		count          int
		expectedDemand map[Provider]int
	}{
		{
			offset: 0, count: 1,
			expectedDemand: map[Provider]int{
				Provider1: 1, Provider2: 1,
			},
		},
		{
			offset: 0, count: 5,
			expectedDemand: map[Provider]int{
				Provider1: 4, Provider2: 3, Provider3: 2,
			},
		},
		{
			offset: 0, count: len(DefaultConfig),
			expectedDemand: map[Provider]int{
				Provider1: 6, Provider2: 6, Provider3: 3,
			},
		},
		{
			offset: 0, count: len(DefaultConfig) * 6,
			expectedDemand: map[Provider]int{
				Provider1: 36, Provider2: 36, Provider3: 18,
			},
		},
		{
			offset: 3, count: len(DefaultConfig) * 6,
			expectedDemand: map[Provider]int{
				Provider1: 36, Provider2: 36, Provider3: 18,
			},
		},
		{
			offset: 15, count: 15,
			expectedDemand: map[Provider]int{
				Provider1: 11, Provider2: 11, Provider3: 6,
			},
		},
	}

	dm := NewDemandManager(DefaultConfig)
	for _, c := range cases {
		t.Run(fmt.Sprintf("count/%d/offset/%d", c.count, c.offset), func(t *testing.T) {
			counts := dm.ProvidersCounts(c.count, c.offset)
			require.EqualValues(t, c.expectedDemand, counts)
			total := 0
			for _, v := range counts {
				require.LessOrEqual(t, v, c.count)
				total += v
			}
			require.LessOrEqual(t, total, c.count*2)
			require.GreaterOrEqual(t, total, c.count)
		})
	}
}
