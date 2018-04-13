package catalog

import (
	"log"
	"strings"
	"github.com/andrew00x/gomovies/api"
	"github.com/andrew00x/gomovies/config"
)

type Index interface {
	Add(m api.Movie)
	Find(title string) []int
}

type IndexFactory func(*config.Config) (Index, error)

func CreateIndex(conf *config.Config) (Index, error) {
	return indexFactory(conf)
}

var indexFactory IndexFactory

func init() {
	indexFactory = func(_ *config.Config) (Index, error) {
		return &SimpleIndex{make(map[string]int)}, nil
	}
}

type SimpleIndex struct {
	idx map[string]int
}

func (i *SimpleIndex) Add(m api.Movie) {
	i.idx[strings.ToLower(m.Title)] = m.Id
	log.Printf("Add file '%s' to index\n", m.Path)
}

func (i *SimpleIndex) Find(title string) []int {
	lower := strings.ToLower(title)
	var result []int
	for key, id := range i.idx {
		if strings.Contains(key, lower) {
			result = append(result, id)
		}
	}
	return result
}
