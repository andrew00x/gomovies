package service

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/andrew00x/gomovies/pkg/api"
	"github.com/andrew00x/gomovies/pkg/config"
	"github.com/andrew00x/gomovies/pkg/tmdb"
	"github.com/andrew00x/gomovies/pkg/util"
)

type TMDbService struct {
	conf                 *config.Config
	detailsCache         util.Cache
	tmDbConf             tmdb.Configuration
	tmDbConfUpdateTicker *time.Ticker
}

func CreateTMDbService(conf *config.Config) (*TMDbService, error) {
	srv := TMDbService{conf: conf, detailsCache: util.CreateCache()}
	err := srv.start()
	return &srv, err
}

func (srv *TMDbService) MovieDetails(tmDbId int, load bool) (md *api.MovieDetails, err error) {
	var res interface{}
	if load {
		res, err = srv.detailsCache.GetOrLoad(tmDbId, func(k util.Key) (interface{}, error) {
			return tmdb.GetTmDbInstance(srv.conf.TMDbApiKey).GetMovie(k.(int))
		})
	} else {
		res, err = srv.detailsCache.Get(tmDbId)
	}
	if res != nil {
		tmDbMovie := res.(tmdb.MovieDetails)
		if err == nil {
			md = &api.MovieDetails{}
			md.Budget = tmDbMovie.Budget
			md.Companies = companyNames(tmDbMovie.ProductionCompanies)
			md.Countries = countryNames(tmDbMovie.ProductionCountries)
			md.Genres = genreNames(tmDbMovie.Genres)
			md.OriginalTitle = tmDbMovie.OriginalTitle
			md.Overview = tmDbMovie.Overview
			md.PosterSmallUrl = fmt.Sprintf("%s%s%s", srv.tmDbConf.Images.BaseUrl, srv.conf.TMDbPosterSmall, tmDbMovie.PosterPath)
			md.PosterLargeUrl = fmt.Sprintf("%s%s%s", srv.tmDbConf.Images.BaseUrl, srv.conf.TMDbPosterLarge, tmDbMovie.PosterPath)
			md.ReleaseDate = tmDbMovie.ReleaseDate
			md.Revenue = tmDbMovie.Revenue
			md.TagLine = tmDbMovie.TagLine
			md.TMDbId = tmDbMovie.Id
		}
	}
	return
}

func (srv *TMDbService) SearchDetails(query string) ([]api.MovieDetails, error) {
	result, err := tmdb.GetTmDbInstance(srv.conf.TMDbApiKey).SearchMovies(query)
	if err == nil {
		details := make([]api.MovieDetails, 0, len(result))
		for _, tmDbMovie := range result {
			md := api.MovieDetails{}
			md.OriginalTitle = tmDbMovie.OriginalTitle
			md.Overview = tmDbMovie.Overview
			md.PosterSmallUrl = fmt.Sprintf("%s%s%s", srv.tmDbConf.Images.BaseUrl, srv.conf.TMDbPosterSmall, tmDbMovie.PosterPath)
			md.PosterLargeUrl = fmt.Sprintf("%s%s%s", srv.tmDbConf.Images.BaseUrl, srv.conf.TMDbPosterLarge, tmDbMovie.PosterPath)
			md.ReleaseDate = tmDbMovie.ReleaseDate
			md.TMDbId = tmDbMovie.Id
			details = append(details, md)
		}
		return details, nil
	}
	return nil, err
}

func (srv *TMDbService) start() (err error) {
	srv.tmDbConf, err = tmdb.GetTmDbInstance(srv.conf.TMDbApiKey).GetConfiguration()
	if err == nil {
		srv.tmDbConfUpdateTicker = time.NewTicker(48 * time.Hour)
		go func() {
			for range srv.tmDbConfUpdateTicker.C {
				if c, err := tmdb.GetTmDbInstance(srv.conf.TMDbApiKey).GetConfiguration(); err != nil {
					log.WithFields(log.Fields{"err": err}).Error("Failed update TMDb configuration")
				} else {
					srv.tmDbConf = c
					log.Info("Successfully update configuration")
				}
			}
		}()
	}
	return
}

func (srv *TMDbService) Stop() {
	if srv.tmDbConfUpdateTicker != nil {
		srv.tmDbConfUpdateTicker.Stop()
	}
	return
}

func companyNames(companies []tmdb.Company) []string {
	names := make([]string, 0, len(companies))
	for _, c := range companies {
		names = append(names, c.Name)
	}
	return names
}

func countryNames(countries []tmdb.Country) []string {
	names := make([]string, 0, len(countries))
	for _, c := range countries {
		names = append(names, c.Name)
	}
	return names
}

func genreNames(genres []tmdb.Genre) []string {
	names := make([]string, 0, len(genres))
	for _, c := range genres {
		names = append(names, c.Name)
	}
	return names
}
