package tmdb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type MovieSearchResult struct {
	Page         int          `json:"page"`
	TotalPages   int          `json:"total_pages"`
	TotalResults int          `json:"total_results"`
	Results      []MovieShort `json:"results"`
}

type MovieShort struct {
	Id               int    `json:"id"`
	Title            string `json:"title"`
	PosterPath       string `json:"poster_path"`
	BackdropPath     string `json:"backdrop_path"`
	OriginalTitle    string `json:"original_title"`
	OriginalLanguage string `json:"original_language"`
	Overview         string `json:"overview"`
	ReleaseDate      string `json:"release_date"`
}

type MovieDetails struct {
	Id                  int       `json:"id"`
	Title               string    `json:"title"`
	OriginalTitle       string    `json:"original_title"`
	TagLine             string    `json:"tagline"`
	PosterPath          string    `json:"poster_path"`
	BackdropPath        string    `json:"backdrop_path"`
	ReleaseDate         string    `json:"release_date"`
	Revenue             int64     `json:"revenue"`
	ProductionCountries []Country `json:"production_countries"`
	ProductionCompanies []Company `json:"production_companies"`
	Budget              int64     `json:"budget"`
	OriginalLanguage    string    `json:"original_language"`
	Overview            string    `json:"overview"`
	Genres              []Genre   `json:"genres"`
	ImdbId              string    `json:"imdb_id"`
	Runtime             int       `json:"runtime"`
}

type GenreResult struct {
	Genres []Genre `json:"genres"`
}

type Genre struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Country struct {
	Code string `json:"iso_3166_1"`
	Name string `json:"name"`
}

type Company struct {
	Id       int    `json:"id"`
	LogoPath string `json:"logo_path"`
	Name     string `json:"name"`
	Country  string `json:"origin_country"`
}

type ErrorStatus struct {
	Code    int    `json:"status_code"`
	Message string `json:"status_message"`
}

type Config struct {
	Images Images `json:"images"`
}

type Images struct {
	BaseUrl       string   `json:"base_url"`
	SecureBaseUrl string   `json:"secure_base_url"`
	BackdropSizes []string `json:"backdrop_sizes"`
	LogoSizes     []string `json:"logo_sizes"`
	PosterSizes   []string `json:"poster_sizes"`
	ProfileSizes  []string `json:"profile_sizes"`
	StillSizes    []string `json:"still_sizes"`
}

const baseUrl = "https://api.themoviedb.org/3"
const remainingReqLimitHeader = "X-RateLimit-Remaining"
const rateLimitEndsHeader = "X-RateLimit-Reset"

var clientFactory apiClientFactory

func init() {
	clientFactory = func() apiClient { return &defaultApiClient{} }
}

type TmDb struct {
	mu        sync.Mutex
	rateTimer time.Time
	apiKey    string
	client    apiClient
}

type apiClientFactory func() apiClient
type apiClient interface {
	get(reqUrl string) (*http.Response, error)
}
type defaultApiClient struct{}

func (*defaultApiClient) get(reqUrl string) (*http.Response, error) {
	return http.Get(reqUrl)
}

var tmDb *TmDb
var once sync.Once

func GetTmDbInstance(apiKey string) *TmDb {
	once.Do(func() {
		tmDb = &TmDb{apiKey: apiKey, client: clientFactory()}
	})
	return tmDb
}

func (tmdb *TmDb) GetGenres(lang string) ([]Genre, error) {
	reqUrl := fmt.Sprintf("%s/genre/movie/list?api_key=%s&language=%s", baseUrl, tmdb.apiKey, lang)
	genres := GenreResult{}
	_, err := tmdb.request(reqUrl, &genres)
	return genres.Genres, err
}

func (tmdb *TmDb) GetConfiguration() (Config, error) {
	reqUrl := fmt.Sprintf("%s/configuration?api_key=%s", baseUrl, tmdb.apiKey)
	config := Config{}
	_, err := tmdb.request(reqUrl, &config)
	return config, err
}

func (tmdb *TmDb) SearchMovies(query, lang string) ([]MovieShort, error) {
	reqUrlFormat := "%s/search/movie?api_key=%s&query=%s&page=%d&language=%s"
	reqUrl := fmt.Sprintf(reqUrlFormat, baseUrl, tmdb.apiKey, url.QueryEscape(query), 1, lang)
	result := MovieSearchResult{}
	_, err := tmdb.request(reqUrl, &result)
	all := make([]MovieShort, 0, result.TotalResults)
	for _, m := range result.Results {
		all = append(all, m)
	}
	totalPages := result.TotalPages
	if totalPages > 1 {
		for page := 2; err == nil && page <= totalPages; page++ {
			reqUrl = fmt.Sprintf(reqUrlFormat, baseUrl, tmdb.apiKey, url.QueryEscape(query), page, lang)
			_, err = tmdb.request(reqUrl, &result)
			for _, m := range result.Results {
				all = append(all, m)
			}
		}
	}
	return all, err
}

func (tmdb *TmDb) GetMovie(id int, lang string) (MovieDetails, error) {
	reqUrl := fmt.Sprintf("%s/movie/%d?api_key=%s&language=%s", baseUrl, id, tmdb.apiKey, lang)
	mov := MovieDetails{}
	_, err := tmdb.request(reqUrl, &mov)
	if err != nil {
		log.WithFields(log.Fields{"movie_id": id, "lang": lang, "err": err}).Error("Error occurred while retrieving movie details from TMDb")
	} else {
		log.WithFields(log.Fields{"movie_id": id, "lang": lang, "title": mov.Title}).Info("Found details in TMDb")
	}
	return mov, err
}

func (tmdb *TmDb) request(reqUrl string, payload interface{}) (interface{}, error) {
	tmdb.mu.Lock()
	defer tmdb.mu.Unlock()

	now := time.Now()
	if tmdb.rateTimer.After(now) {
		<-time.After(tmdb.rateTimer.Sub(now))
	}

	resp, err := tmdb.client.get(reqUrl)
	if err != nil {
		return nil, err
	}
	defer func() {
		if clsErr := resp.Body.Close(); clsErr != nil {
			log.Warn(clsErr)
		}
	}()

	if resp.Header.Get(remainingReqLimitHeader) == "0" {
		rateLimitEnds, err := strconv.ParseInt(resp.Header.Get(rateLimitEndsHeader), 10, 64)
		if err == nil {
			tmdb.rateTimer = time.Unix(1+rateLimitEnds, 0)
		}
	}

	parser := json.NewDecoder(resp.Body)
	if resp.StatusCode/100 == 2 {
		err = parser.Decode(payload)
		if err != nil {
			return nil, err
		}
		return payload, nil
	}

	errorStatus := ErrorStatus{}
	err = parser.Decode(&errorStatus)
	if err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("error: %s, code: %d", errorStatus.Message, errorStatus.Code)
}
