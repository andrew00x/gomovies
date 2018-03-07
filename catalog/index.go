package catalog

import (
	"log"
	"strings"
)

type index struct {
	idx map[string]int
}

func newIndex(size int) *index {
	return &index{make(map[string]int, size)}
}

func (i *index) add(f MovieFile) {
	i.idx[strings.ToLower(f.Title)] = f.Id
	log.Printf("Add file '%s' to index\n", f.Path)
}

func (i *index) find(title string) []int {
	lower := strings.ToLower(title)
	var result []int
	for key, id := range i.idx {
		if strings.Contains(key, lower) {
			result = append(result, id)
		}
	}
	return result
}
