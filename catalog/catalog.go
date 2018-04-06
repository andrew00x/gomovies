package catalog

import "github.com/andrew00x/gomovies/config"

type MovieFile struct {
	Id        int    `json:"id"`
	Title     string `json:"title"`
	Path      string `json:"path"`
	DriveName string `json:"drive"`
	TMDbId    string `json:"tmdb_id"`
}

type CatalogFactory func(*config.Config) (Catalog, error)

var catalogFactory CatalogFactory

func CreateCatalog(conf *config.Config) (Catalog, error) {
	return catalogFactory(conf)
}

type Catalog interface {
	Get(id int) *MovieFile
	Find(title string) []MovieFile
	All() []MovieFile
	Save() error
	Refresh(conf *config.Config) error
}
