package apigee

import (
	"fmt"
	"time"

	"github.com/Axway/agent-sdk/pkg/cache"
)

type agentCache struct {
	cache              cache.Cache
	specEndpointToKeys map[string][]specCacheItem
}

type specCacheItem struct {
	Hash        uint64
	ContentPath string
	ModDate     time.Time
}

func newAgentCache() *agentCache {
	return &agentCache{
		cache:              cache.New(),
		specEndpointToKeys: make(map[string][]specCacheItem),
	}
}

func (a *agentCache) AddSpecToCache(id, path string, contentHash uint64, modDate time.Time, endpoints ...string) {
	item := specCacheItem{
		Hash:        contentHash,
		ContentPath: path,
		ModDate:     modDate,
	}

	a.cache.SetWithSecondaryKey(id, path, item)
	for _, ep := range endpoints {
		if _, found := a.specEndpointToKeys[ep]; !found {
			a.specEndpointToKeys[ep] = []specCacheItem{}
		}
		a.specEndpointToKeys[ep] = append(a.specEndpointToKeys[ep], item)
	}
}

func (a *agentCache) GetSpecWithPath(path string) (uint64, error) {
	data, err := a.cache.GetBySecondaryKey(path)
	if err != nil {
		return 0, err
	}

	contentHash := data.(uint64)
	return contentHash, nil
}

// GetSpecPathWithEndpoint - returns the lat modified spec found with this endpoint
func (a *agentCache) GetSpecPathWithEndpoint(endpoint string) (string, error) {
	items, found := a.specEndpointToKeys[endpoint]
	if !found {
		return "", fmt.Errorf("no spec found for endpoint: %s", endpoint)
	}

	latest := specCacheItem{}
	for _, item := range items {
		if item.ModDate.After(latest.ModDate) {
			latest = item
		}
	}

	return latest.ContentPath, nil
}
