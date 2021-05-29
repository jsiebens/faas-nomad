package services

import (
	"fmt"
	"net/url"
	"time"

	"github.com/hashicorp/consul-template/dependency"
	"github.com/hashicorp/consul-template/watch"
	"github.com/jsiebens/faas-nomad/pkg/types"
	"github.com/patrickmn/go-cache"
)

type ServiceResolver interface {
	Resolve(functionName string) (url.URL, error)
	RemoveCacheItem(functionName string)
}

type consulResolver struct {
	clientSet *dependency.ClientSet
	watcher   *watch.Watcher
	cache     *cache.Cache
	prefix    string
}

type cacheItem struct {
	serviceQuery dependency.Dependency
	addresses    []url.URL
}

func (c cacheItem) next() url.URL {
	return c.addresses[randIntn(len(c.addresses))]
}

func NewConsulResolver(config *types.ProviderConfig) (ServiceResolver, error) {
	clientSet := dependency.NewClientSet()
	err := clientSet.CreateConsulClient(&dependency.CreateConsulClientInput{
		Address: config.Consul.Addr,
	})

	if err != nil {
		return nil, err
	}

	watcher, _ := watch.NewWatcher(&watch.NewWatcherInput{
		Clients:  clientSet,
		MaxStale: 10000 * time.Millisecond,
	})

	pc := cache.New(5*time.Minute, 10*time.Minute)

	resolver := &consulResolver{
		clientSet: clientSet,
		watcher:   watcher,
		cache:     pc,
		prefix:    config.Scheduling.JobPrefix,
	}

	go resolver.watch()

	return resolver, nil
}

func (sr *consulResolver) Resolve(functionName string) (url.URL, error) {
	return sr.resolveInternal(fmt.Sprintf("%s%s", sr.prefix, functionName))
}

func (sr *consulResolver) resolveInternal(function string) (url.URL, error) {
	if val, ok := sr.cache.Get(getCacheKey(function)); ok {
		return val.(*cacheItem).next(), nil
	}

	q, err := dependency.NewCatalogServiceQuery(function)
	if err != nil {
		return url.URL{}, err
	}

	s, _, err := q.Fetch(sr.clientSet, nil)
	if err != nil {
		return url.URL{}, err
	}
	sr.watcher.Add(q)

	cs := s.([]*dependency.CatalogService)
	item := sr.updateCatalog(q, cs)

	return item.next(), nil
}

func (sr *consulResolver) RemoveCacheItem(function string) {
	key := getCacheKey(function)
	if d, ok := sr.cache.Get(key); ok {
		sr.watcher.Remove(d.(*cacheItem).serviceQuery)
		sr.cache.Delete(key)
	}
}

func (sr *consulResolver) watch() {
	for cs := range sr.watcher.DataCh() {
		sr.updateCatalog(
			cs.Dependency(),
			cs.Data().([]*dependency.CatalogService),
		)
	}
}

func (sr *consulResolver) updateCatalog(dep dependency.Dependency, cs []*dependency.CatalogService) *cacheItem {
	addresses := make([]url.URL, 0)

	if len(cs) < 1 {
		return sr.upsertCache(dep, addresses)
	}

	for _, addr := range cs {
		addresses = append(
			addresses,
			toUrl(fmt.Sprintf("http://%v:%v", addr.ServiceAddress, addr.ServicePort)),
		)
	}

	return sr.upsertCache(dep, addresses)
}

func (sr *consulResolver) upsertCache(dep dependency.Dependency, value []url.URL) *cacheItem {
	if ci, ok := sr.cache.Get(dep.String()); ok {
		item := ci.(*cacheItem)
		item.addresses = value
		sr.cache.Set(dep.String(), ci, cache.NoExpiration)

		return item
	}

	item := &cacheItem{
		addresses:    value,
		serviceQuery: dep,
	}
	sr.cache.Set(dep.String(), item, cache.NoExpiration)
	return item
}

func getCacheKey(function string) string {
	return fmt.Sprintf("catalog.service(%s)", function)
}

func toUrl(address string) url.URL {
	parse, _ := url.Parse(address)
	return *parse
}

var randIntn = func(n int) int {
	if n == 0 {
		return 0
	}
	return int(time.Now().UnixNano()/int64(time.Microsecond)) % n
}
