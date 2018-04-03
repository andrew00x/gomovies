package catalog

import "github.com/andrew00x/gomovies/config"

type Factory func(*config.Config) (Catalog, error)

var factory Factory

func Create(conf *config.Config) (Catalog, error) {
	c, err := factory(conf)
	if err != nil {
		return nil, err
	}
	return c, nil
}

type Catalog interface {
	Get(id int) *MovieFile
	Find(title string) []MovieFile
	All() []MovieFile
	Save() error
	Refresh(conf *config.Config) error
}
