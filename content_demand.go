package main

type ContentDemand struct {
	demandLookup []map[Provider]int
}

// NewContentDemand calculates
func NewContentDemand(config ContentMix) ContentDemand {
	demands := make([]map[Provider]int, len(config))
	currentDemand := make(map[Provider]int)
	for i, contentConfig := range config {
		currentDemand[contentConfig.Type]++
		if contentConfig.Fallback != nil {
			currentDemand[*contentConfig.Fallback]++
		}

		demands[i] = make(map[Provider]int)
		for k, v := range currentDemand {
			demands[i][k] = v
		}
	}
	return ContentDemand{
		demandLookup: demands,
	}
}

func (cd *ContentDemand) ProvidersCounts(count, offset int) map[Provider]int {
	lookupLen := len(cd.demandLookup)
	offset %= lookupLen
	counts := make(map[Provider]int)
	i := (offset+count)%lookupLen - 1
	if i >= 0 {
		for k, v := range cd.demandLookup[(offset+count)%lookupLen] {
			counts[k] += v
		}
	}

	i = (offset + count) / lookupLen
	if i > 0 {
		for k, v := range cd.demandLookup[lookupLen-1] {
			counts[k] += v * i
		}
	}

	if offset != 0 {
		for k, v := range cd.demandLookup[offset-1] {
			counts[k] -= v
		}
	}

	return counts
}
