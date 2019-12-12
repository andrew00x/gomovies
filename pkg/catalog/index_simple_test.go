package catalog

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddMovieInIndex(t *testing.T) {
	index := SimpleIndex{make(map[string]map[int]void)}
	movies := []struct {
		id     int
		title  string
		genres []string
	}{
		{id: 1, title: "Green Mile.mkv", genres: []string{"crime", "drama"} },
		{id: 2, title: "Fight Club.mkv", genres: []string{"drama"}},
	}
	for _, m := range movies {
		index.Add(m.title, m.id)
		for _, genre := range m.genres {
			index.Add(genre, m.id)
		}
	}
	expected := map[string]map[int]void{
		"green mile.mkv": {1: emptyValue},
		"fight club.mkv": {2: emptyValue},
		"drama": {1: emptyValue, 2: emptyValue},
		"crime": {1: emptyValue},
	}
	assert.Equal(t, expected, index.idx)
}

func TestFindMovieInIndex(t *testing.T) {
	index := SimpleIndex{make(map[string]map[int]void)}
	movies := []struct {
		id    int
		title string
	}{
		{id: 1, title: "Back to the Future 1.mkv"},
		{id: 2, title: "Back to the Future 2.mkv"},
		{id: 3, title: "The Replacements.mkv"},
	}
	for _, m := range movies {
		index.Add(m.title, m.id)
	}
	expected := []int{1, 2}
	result := index.Find("Futur")
	sort.Ints(result)
	assert.Equal(t, expected, result)
}

func TestFindMovieInIndexIgnoringCase(t *testing.T) {
	index := SimpleIndex{make(map[string]map[int]void)}
	movies := []struct {
		id    int
		title string
	}{
		{id: 1, title: "Brave Heart.mkv"},
		{id: 2, title: "Rush Hour 1.mkv"},
		{id: 3, title: "Rush Hour 2.mkv"},
	}
	for _, m := range movies {
		index.Add(m.title, m.id)
	}
	expected := []int{2, 3}
	result := index.Find("hoUr")
	sort.Ints(result)
	assert.Equal(t, expected, result)
}
