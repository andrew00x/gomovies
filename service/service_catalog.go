package service

import (
	"sort"
	"github.com/andrew00x/gomovies/api"
	"github.com/andrew00x/gomovies/catalog"
	"github.com/andrew00x/gomovies/config"
	"github.com/andrew00x/gomovies/file"
)

type CatalogService struct {
	ctl catalog.Catalog
}

type ByName []api.Movie

func (m ByName) Len() int           { return len(m) }
func (m ByName) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m ByName) Less(i, j int) bool { return m[i].Title < m[j].Title }

func CreateCatalogService(ctl catalog.Catalog) *CatalogService {
	return &CatalogService{ctl}
}

func (srv *CatalogService) Stop() error {
	return srv.ctl.Save()
}

func (srv *CatalogService) All() []api.Movie {
	m := toMovies(srv.ctl.All())
	sort.Sort(ByName(m))
	return m
}

func (srv *CatalogService) Find(title string) []api.Movie {
	m := toMovies(srv.ctl.Find(title))
	sort.Sort(ByName(m))
	return m
}

func (srv *CatalogService) Refresh(config *config.Config) error {
	return srv.ctl.Refresh(config)
}

func toMovies(files []catalog.MovieFile) []api.Movie {
	movies := make([] api.Movie, 0, len(files))
	for _, f := range files {
		exists, err := file.Exists(f.Path)
		m := api.Movie{Id: f.Id, Title: f.Title, Path: f.Path, DriveName: f.DriveName, Available: exists && err == nil}
		movies = append(movies, m)
	}
	return movies
}
