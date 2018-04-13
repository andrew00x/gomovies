package catalog

import (
	"github.com/andrew00x/gomovies/api"
	"github.com/andrew00x/gomovies/config"
)

type CatalogFactory func(*config.Config) (Catalog, error)

var catalogFactory CatalogFactory

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
}
