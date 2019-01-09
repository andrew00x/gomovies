package util

import (
	"errors"
	"sync"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestGetNonexistentItem(t *testing.T) {
	c := Cache{items: map[Key]*item{}}
	k := 1
	v, e := c.Get(k)
	assert.Nil(t, v)
	assert.Nil(t, e)
}

func TestGetItem(t *testing.T) {
	c := Cache{items: map[Key]*item{}}
	value := "item_1"
	k := 1
	c.items[k] = &item{value: value}
	v, e := c.Get(k)
	assert.Equal(t, value, v)
	assert.Nil(t, e)
}

func TestGetError(t *testing.T) {
	c := Cache{items: map[Key]*item{}}
	err := errors.New("error")
	k := 1
	c.items[k] = &item{err: err}
	v, e := c.Get(k)
	assert.Nil(t, v)
	assert.Equal(t, err, e)
}

func TestLoadItem(t *testing.T) {
	c := Cache{items: map[Key]*item{}}
	value := "loaded_item_1"
	loader := func(k Key) (interface{}, error) {
		if k == 1 {
			return value, nil
		}
		return nil, nil
	}
	v, e := c.GetOrLoad(1, loader)
	assert.Equal(t, value, v)
	assert.Nil(t, e)
}

func TestLoadItemOnlyOnce(t *testing.T) {
	c := Cache{items: map[Key]*item{}}
	counter := 0
	var mu sync.Mutex
	loader := func(k Key) (interface{}, error) {
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
			c.GetOrLoad(1, loader)
			g.Done()
		}()
	}
	g.Wait()
	assert.Equal(t, 1, counter)
}

func TestDeleteItem(t *testing.T) {
	m := map[Key]*item{}
	k := 1
	m[k] = &item{value: "value"}
	c := Cache{items: m}
	d := c.Delete(k)
	assert.True(t, d)
	_, found := m[k]
	assert.False(t, found)
}
