package tmdb

import (
	"errors"
	"sync"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestGetNonexistentItem(t *testing.T) {
	c := cache{items: map[key]item{}}
	k := 1
	v, e := c.get(k)
	assert.Nil(t, v)
	assert.Nil(t, e)
}

func TestGetItem(t *testing.T) {
	c := cache{items: map[key]item{}}
	value := "item_1"
	k := 1
	c.items[k] = item{value: value}
	v, e := c.get(k)
	assert.Equal(t, value, v)
	assert.Nil(t, e)
}

func TestGetError(t *testing.T) {
	c := cache{items: map[key]item{}}
	err := errors.New("error")
	k := 1
	c.items[k] = item{err: err}
	v, e := c.get(k)
	assert.Nil(t, v)
	assert.Equal(t, err, e)
}

func TestLoadItem(t *testing.T) {
	c := cache{items: map[key]item{}}
	value := "loaded_item_1"
	loader := func(k key) (interface{}, error) {
		if k == 1 {
			return value, nil
		}
		return nil, nil
	}
	v, e := c.getOrLoad(1, loader)
	assert.Equal(t, value, v)
	assert.Nil(t, e)
}

func TestLoadItemOnlyOnce(t *testing.T) {
	c := cache{items: map[key]item{}}
	counter := 0
	var mu sync.Mutex
	loader := func(k key) (interface{}, error) {
		mu.Lock()
		counter++
		mu.Unlock()
		return nil, nil
	}
	g := sync.WaitGroup{}
	callNum := 10
	g.Add(callNum)
	for i := 0; i < callNum; i++ {
		go func() {
			c.getOrLoad(1, loader)
			g.Done()
		}()
	}
	g.Wait()
	assert.Equal(t, 1, counter)
}

func TestDeleteItem(t *testing.T) {
	m := map[key]item{}
	k := 1
	m[k] = item{value: "value"}
	c := cache{items: m}
	d := c.delete(k)
	assert.True(t, d)
	_, found := m[k]
	assert.False(t, found)
}
