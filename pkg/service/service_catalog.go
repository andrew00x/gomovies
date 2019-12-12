package service

import (
	"sort"

	"github.com/andrew00x/gomovies/pkg/api"
	"github.com/andrew00x/gomovies/pkg/catalog"
	"github.com/andrew00x/gomovies/pkg/config"
)

type CatalogService struct {
	ctl  catalog.Catalog
	conf *config.Config
}

type ByName []api.Movie

func (m ByName) Len() int           { return len(m) }
func (m ByName) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m ByName) Less(i, j int) bool { return m[i].Title < m[j].Title }

func CreateCatalogService(conf *config.Config) (*CatalogService, error) {
	ctl, err := catalog.CreateCatalog(conf)
	if err != nil {
		return nil, err
	}
	return createCatalogService(ctl, conf), nil
}

func createCatalogService(ctl catalog.Catalog, conf *config.Config) *CatalogService {
	return &CatalogService{ctl: ctl, conf: conf}
}

func (srv *CatalogService) All() []api.Movie {
	res := srv.ctl.All()
	sort.Sort(ByName(res))
	return res
}

func (srv *CatalogService) Get(id int) (api.Movie, bool) {
	return srv.ctl.Get(id)
}

func (srv *CatalogService) Find(title string) []api.Movie {
	res := srv.ctl.Find(title)
	sort.Sort(ByName(res))
	return res
}

func (srv *CatalogService) Save() error {
	return srv.ctl.Save()
}

func (srv *CatalogService) Refresh() error {
	return srv.ctl.Refresh()
}

func (srv *CatalogService) Update(u api.Movie) (api.Movie, error) {
	return srv.ctl.Update(u)
}

func (srv *CatalogService) AddTag(tag string, id int) error {
	return srv.ctl.AddTag(tag, id)
}
