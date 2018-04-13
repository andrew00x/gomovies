package service

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"github.com/andrew00x/gomovies/api"
	"github.com/andrew00x/gomovies/config"
	"github.com/andrew00x/gomovies/tmdb"
	"github.com/andrew00x/gomovies/util"
)

type TMDbService struct {
	conf         *config.Config
	detailsCache *util.Cache
	tmDbConf     tmdb.Configuration
}

func CreateTMDbService(conf *config.Config) (*TMDbService, error) {
	srv := TMDbService{conf: conf, detailsCache: util.CreateCache()}
	err := srv.start()
	return &srv, err
}

func (srv *TMDbService) MovieDetails(tmDbId int) (md api.MovieDetails, err error) {
	res, err := srv.detailsCache.GetOrLoad(tmDbId, func(k util.Key) (interface{}, error) {
		return tmdb.GetTmDbInstance(srv.conf.TMDbApiKey).GetMovie(k.(int))
	})
	tmDbMovie := res.(tmdb.MovieDetails)
	if err == nil {
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
		ticker := time.NewTicker(48 * time.Hour)
		go func() {
			for t := range ticker.C {
				if c, err := tmdb.GetTmDbInstance(srv.conf.TMDbApiKey).GetConfiguration(); err != nil {
					log.Printf("failed update TMDb configuration at: %s, error: %v\n", t.Format(time.ANSIC), err)
				} else {
					srv.tmDbConf = c
					log.Printf("sucessfully update configuration at: %s\n", t.Format(time.ANSIC))
				}
			}
		}()

		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
		go func() {
			<- quit
			ticker.Stop()
		}()
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
