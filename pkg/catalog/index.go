package catalog

import (
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/andrew00x/gomovies/pkg/config"
)

type Index interface {
	Add(tag string, id int)
	Find(tag string) []int
}

type IndexFactory func(*config.Config) (Index, error)

type void struct{}

var emptyValue void

func CreateIndex(conf *config.Config) (Index, error) {
	return indexFactory(conf)
}

var indexFactory IndexFactory

func init() {
	indexFactory = func(_ *config.Config) (Index, error) {
		return &SimpleIndex{make(map[string]map[int]void)}, nil
	}
}

type SimpleIndex struct {
	idx map[string]map[int]void
}

func (i *SimpleIndex) Add(tag string, id int) {
	lower := strings.ToLower(tag)
	if _, ok := i.idx[lower]; !ok {
		i.idx[lower] = make(map[int]void)
	}
	if _, ok := i.idx[lower][id]; !ok {
		i.idx[lower][id] = emptyValue
		log.WithFields(log.Fields{"tag": tag, "id": id}).Debug("Add tag to index")
	}
}

func (i *SimpleIndex) Find(tag string) []int {
	lower := strings.ToLower(tag)
	result := map[int]void{}
	for key, ids := range i.idx {
		if strings.Contains(key, lower) {
			for id := range ids {
				if _, ok := result[id]; !ok {
					result[id] = emptyValue
				}
			}
		}
	}
	keys := make([]int, 0, len(result))
	for id := range result {
		keys = append(keys, id)
	}
	return keys
}
