package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnqueue(t *testing.T) {
	q := PlayQueue{}
	q.Enqueue("/a.avi")
	q.Enqueue("/b.avi")
	assert.Equal(t, []string{"/a.avi", "/b.avi"}, q.arr)
}

func TestPop(t *testing.T) {
	q := PlayQueue{arr: []string{"/a.mkv", "/b.mkv"}}
	p := q.Pop()
	assert.Equal(t, "/a.mkv", p)
	assert.Equal(t, []string{"/b.mkv"}, q.arr)
}

func TestDequeue(t *testing.T) {
	q := PlayQueue{arr: []string{"/a.mkv", "/b.mkv", "/c.mkv"}}
	q.Dequeue(1)
	assert.Equal(t, []string{"/a.mkv", "/c.mkv"}, q.arr)
}

func TestPopWhenEmpty(t *testing.T) {
	q := PlayQueue{}
	p := q.Pop()
	assert.Equal(t, "", p)
}

func TestEmptyWhenEmpty(t *testing.T) {
	q := PlayQueue{}
	assert.Equal(t, true, q.Empty())
}

func TestEmpty(t *testing.T) {
	q := PlayQueue{arr: []string{"/a.avi"}}
	assert.Equal(t, false, q.Empty())
}

func TestAll(t *testing.T) {
	q := PlayQueue{arr: []string{"/a.avi", "/b.avi"}}
	assert.Equal(t, []string{"/a.avi", "/b.avi"}, q.All())
}

func TestAllWhenEmpty(t *testing.T) {
	q := PlayQueue{}
	assert.Equal(t, []string{}, q.All())
}
