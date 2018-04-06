package catalog

import (
	"testing"
	"reflect"
	"sort"
)

func TestAddMovieInIndex(t *testing.T) {
	index := SimpleIndex{make(map[string]int)}
	movies := []MovieFile{
		{Id: 1, Title: "Green Mile.mkv", Path: "/movies/Green Mile.mkv"},
		{Id: 2, Title: "Fight Club.mkv", Path: "/movies/Fight Club.mkv"},
	}
	for _, m := range movies {
		index.Add(m)
	}
	expected := map[string]int{
		"green mile.mkv": 1,
		"fight club.mkv": 2,
	}
	if !reflect.DeepEqual(expected, index.idx) {
		t.Fatalf("expected index is: %+v, but actuals is: %+v", expected, index.idx)
	}
}

func TestFindMovieInIndex(t *testing.T) {
	index := SimpleIndex{make(map[string]int)}
	movies := []MovieFile{
		{Id: 1, Title: "Back to the Future 1.mkv", Path: "/movies/Back to the Future 1.mkv"},
		{Id: 2, Title: "Back to the Future 2.mkv", Path: "/movies/Back to the Future 2.mkv"},
		{Id: 3, Title: "The Replacements.mkv", Path: "/movies/The Replacements.mkv"},
	}
	for _, m := range movies {
		index.Add(m)
	}
	expected := []int{1, 2}
	result := index.Find("Futur")
	sort.Ints(result)
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("expected index is: %+v, but actuals is: %+v", expected, result)
	}
}

func TestFindMovieInIndexIgnoringCase(t *testing.T) {
	index := SimpleIndex{make(map[string]int)}
	movies := []MovieFile{
		{Id: 1, Title: "Brave Heart.mkv", Path: "/movies/Brave Heart.mkv"},
		{Id: 2, Title: "Rush Hour 1.mkv", Path: "/movies/Rush Hour 1.mkv"},
		{Id: 3, Title: "Rush Hour 2.mkv", Path: "/movies/Rush Hour 2.mkv"},
	}
	for _, m := range movies {
		index.Add(m)
	}
	expected := []int{2, 3}
	result := index.Find("hoUr")
	sort.Ints(result)
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("expected index is: %+v, but actuals is: %+v", expected, result)
	}
}
