package catalog

import (
	"sort"
	"testing"

	"github.com/andrew00x/gomovies/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestAddMovieInIndex(t *testing.T) {
	index := SimpleIndex{make(map[string]int)}
	movies := []api.Movie{
		{Id: 1, Title: "Green Mile.mkv", File: "/movies/Green Mile.mkv"},
		{Id: 2, Title: "Fight Club.mkv", File: "/movies/Fight Club.mkv"},
	}
	for _, m := range movies {
		index.Add(m)
	}
	expected := map[string]int{
		"green mile.mkv": 1,
		"fight club.mkv": 2,
	}
	assert.Equal(t, expected, index.idx)
}

func TestFindMovieInIndex(t *testing.T) {
	index := SimpleIndex{make(map[string]int)}
	movies := []api.Movie{
		{Id: 1, Title: "Back to the Future 1.mkv", File: "/movies/Back to the Future 1.mkv"},
		{Id: 2, Title: "Back to the Future 2.mkv", File: "/movies/Back to the Future 2.mkv"},
		{Id: 3, Title: "The Replacements.mkv", File: "/movies/The Replacements.mkv"},
	}
	for _, m := range movies {
		index.Add(m)
	}
	expected := []int{1, 2}
	result := index.Find("Futur")
	sort.Ints(result)
	assert.Equal(t, expected, result)
}

func TestFindMovieInIndexIgnoringCase(t *testing.T) {
	index := SimpleIndex{make(map[string]int)}
	movies := []api.Movie{
		{Id: 1, Title: "Brave Heart.mkv", File: "/movies/Brave Heart.mkv"},
		{Id: 2, Title: "Rush Hour 1.mkv", File: "/movies/Rush Hour 1.mkv"},
		{Id: 3, Title: "Rush Hour 2.mkv", File: "/movies/Rush Hour 2.mkv"},
	}
	for _, m := range movies {
		index.Add(m)
	}
	expected := []int{2, 3}
	result := index.Find("hoUr")
	sort.Ints(result)
	assert.Equal(t, expected, result)
}
