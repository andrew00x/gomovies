package util

import "sync"

type Cache struct {
	mu    sync.Mutex
	items map[Key]*item
}

type Key interface{}

type item struct {
	value  interface{}
	err    error
	loaded chan bool
}

func CreateCache() *Cache {
	return &Cache{items: make(map[Key]*item)}
}

func (c *Cache) Get(k Key) (interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if result, found := c.items[k]; found {
		return result.value, result.err
	}
	return nil, nil
}

func (c *Cache) GetOrLoad(k Key, loader func(k Key) (interface{}, error)) (interface{}, error) {
	c.mu.Lock()
	result, found := c.items[k]
	if found {
		c.mu.Unlock()
		<-result.loaded
	} else {
		result = &item{loaded: make(chan bool)}
		c.items[k] = result
		c.mu.Unlock()

		result.value, result.err = loader(k)
		close(result.loaded)
	}
	return result.value, result.err
}

func (c *Cache) Delete(k Key) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, found := c.items[k]; found {
		delete(c.items, k)
		return true
	}
	return false
}

func (c *Cache) Clean() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[Key]*item)
}
