package service

import (
	"fmt"

	"github.com/andrew00x/gomovies/pkg/api"
	"github.com/andrew00x/gomovies/pkg/config"
	"github.com/andrew00x/gomovies/pkg/details"
	"github.com/andrew00x/gomovies/pkg/tmdb"
	"github.com/andrew00x/gomovies/pkg/util"
)

type DetailsService struct {
	tmdbLoader  *tmdbLoader
	localLoader *localLoader
}

func CreateDetailsService(conf *config.Config) (srv *DetailsService, err error) {
	if conf.TMDbApiKey == "" {
		srv = &DetailsService{
			tmdbLoader: nil,
			localLoader: &localLoader{
				cache: util.CreateCache(),
			},
		}
	} else {
		tmdbConn := tmdb.GetTmDbInstance(conf.TMDbApiKey)
		var tmdbConf tmdb.Config
		tmdbConf, err = tmdbConn.GetConfiguration()
		if err != nil {
			return
		}
		srv = &DetailsService{
			tmdbLoader: &tmdbLoader{
				conf:     conf,
				tmdbConf: tmdbConf,
				tmdbConn: tmdbConn,
				cache:    util.CreateCache(),
			},
			localLoader: &localLoader{
				cache: util.CreateCache(),
			},
		}
	}
	return
}

func (srv *DetailsService) MovieDetails(m api.Movie, lang string, tryLoad bool) (api.MovieDetails, bool, error) {
	if m.TMDbId != 0 && srv.tmdbLoader != nil {
		return srv.tmdbLoader.load(m, lang, tryLoad)
	}
	return srv.localLoader.load(m, lang, tryLoad)
}

func (srv *DetailsService) SearchDetails(query, lang string) ([]api.MovieDetails, error) {
	if srv.tmdbLoader != nil {
		result, err := srv.tmdbLoader.tmdbConn.SearchMovies(query, lang)
		if err == nil {
			movies := make([]api.MovieDetails, 0, len(result))
			for _, tmDbMovie := range result {
				movies = append(movies,
					api.MovieDetails{
						OriginalTitle:  tmDbMovie.OriginalTitle,
						Overview:       tmDbMovie.Overview,
						PosterSmallUrl: fmt.Sprintf("%s%s%s", srv.tmdbLoader.tmdbConf.Images.BaseUrl, srv.tmdbLoader.conf.TMDbPosterSmall, tmDbMovie.PosterPath),
						PosterLargeUrl: fmt.Sprintf("%s%s%s", srv.tmdbLoader.tmdbConf.Images.BaseUrl, srv.tmdbLoader.conf.TMDbPosterLarge, tmDbMovie.PosterPath),
						ReleaseDate:    tmDbMovie.ReleaseDate,
						TMDbId:         tmDbMovie.Id,
					})
			}
			return movies, nil
		}
		return nil, err
	}
	return nil, nil
}

type localDetailsKey struct {
	file string
	lang string
}

type localLoader struct {
	cache util.Cache
}

func (l *localLoader) load(m api.Movie, lang string, tryLoad bool) (md api.MovieDetails, found bool, err error) {
	var v interface{}
	if tryLoad {
		v, err = l.cache.GetOrLoad(localDetailsKey{file: m.File, lang: lang}, func(key util.Key) (interface{}, error) {
			k := key.(localDetailsKey)
			return details.GetDetails(k.file, k.lang)
		})
	} else {
		v, err = l.cache.Get(localDetailsKey{file: m.File, lang: lang})
	}
	if err == nil && v != nil {
		found = true
		md = v.(api.MovieDetails)
	}
	return
}

type tmdbDetailsKey struct {
	id   int
	lang string
}

type tmdbLoader struct {
	conf     *config.Config
	tmdbConf tmdb.Config
	tmdbConn *tmdb.TmDb
	cache    util.Cache
}

func (l *tmdbLoader) load(m api.Movie, lang string, tryLoad bool) (md api.MovieDetails, found bool, err error) {
	var v interface{}
	if tryLoad {
		v, err = l.cache.GetOrLoad(tmdbDetailsKey{id: m.TMDbId, lang: lang}, func(key util.Key) (interface{}, error) {
			k := key.(tmdbDetailsKey)
			return l.tmdbConn.GetMovie(k.id, k.lang)
		})
	} else {
		v, err = l.cache.Get(tmdbDetailsKey{id: m.TMDbId, lang: lang})
	}
	if err == nil && v != nil {
		found = true
		tmDbMovie := v.(tmdb.MovieDetails)
		md = api.MovieDetails{
			Budget:         tmDbMovie.Budget,
			Companies:      companyNames(tmDbMovie.ProductionCompanies),
			Countries:      countryNames(tmDbMovie.ProductionCountries),
			Genres:         genreNames(tmDbMovie.Genres),
			OriginalTitle:  tmDbMovie.OriginalTitle,
			Overview:       tmDbMovie.Overview,
			PosterSmallUrl: fmt.Sprintf("%s%s%s", l.tmdbConf.Images.BaseUrl, l.conf.TMDbPosterSmall, tmDbMovie.PosterPath),
			PosterLargeUrl: fmt.Sprintf("%s%s%s", l.tmdbConf.Images.BaseUrl, l.conf.TMDbPosterLarge, tmDbMovie.PosterPath),
			Runtime:        tmDbMovie.Runtime,
			ReleaseDate:    tmDbMovie.ReleaseDate,
			Revenue:        tmDbMovie.Revenue,
			TagLine:        tmDbMovie.TagLine,
			Title:          tmDbMovie.Title,
			TMDbId:         tmDbMovie.Id,
		}
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
