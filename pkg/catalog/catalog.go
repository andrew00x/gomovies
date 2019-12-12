package catalog

import (
	"github.com/andrew00x/gomovies/pkg/api"
	"github.com/andrew00x/gomovies/pkg/config"
)

type Factory func(*config.Config) (Catalog, error)

var catalogFactory Factory

func CreateCatalog(conf *config.Config) (Catalog, error) {
	return catalogFactory(conf)
}

type Catalog interface {
	All() []api.Movie
	Find(title string) []api.Movie
	Get(id int) (api.Movie, bool)
	Load() error
	Refresh() error
	Save() error
	Update(u api.Movie) (api.Movie, error)
	AddTag(tag string, id int) error
}
