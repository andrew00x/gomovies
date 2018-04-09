package tmdb

import "sync"

type cache struct {
	mu    sync.Mutex
	items map[key]item
}

type key interface{}

type item struct {
	value  interface{}
	err    error
	loaded chan bool
}

func createCache() *cache {
	return &cache{items: make(map[key]item)}
}

func (c *cache) get(k key) (interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if result, found := c.items[k]; found {
		return result.value, result.err
	}
	return nil, nil
}

func (c *cache) getOrLoad(k key, loader func(k key) (interface{}, error)) (interface{}, error) {
	c.mu.Lock()
	result, found := c.items[k]
	if found {
		c.mu.Unlock()
		<-result.loaded
	} else {
		result = item{loaded: make(chan bool)}
		c.items[k] = result
		c.mu.Unlock()

		result.value, result.err = loader(k)
		close(result.loaded)
	}
	return result.value, result.err
}

func (c *cache) delete(k key) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, found := c.items[k]; found {
		delete(c.items, k)
		return true
	}
	return false
}

func (c *cache) clean() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[key]item)
}
