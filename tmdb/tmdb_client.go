package tmdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
)

const baseUrl = "https://api.themoviedb.org/3"

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
	Tagline             string    `json:"tagline"`
	PosterPath          string    `json:"poster_path"`
	BackdropPath        string    `json:"backdrop_path"`
	ReleaseDate         string    `json:"release_date"`
	Revenue             int       `json:"revenue"`
	ProductionCountries []Country `json:"production_countries"`
	ProductionCompanies []Company `json:"production_companies"`
	Budget              int       `json:"budget"`
	OriginalLanguage    string    `json:"original_language"`
	Overview            string    `json:"overview"`
	Genres              []Genre   `json:"genres"`
	ImdbId              string    `json:"imdb_id"`
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

type Configuration struct {
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

var clientFactory apiClientFactory

func init() {
	clientFactory = func() apiClient { return &defaultApiClient{} }
}

type TmDb struct {
	mu                sync.Mutex
	apiKey            string
	client            apiClient
	movieDetailsCache *cache
}

type apiClientFactory func() apiClient
type apiClient interface {
	get(reqUrl string) (*http.Response, error)
}
type defaultApiClient struct{}

func (*defaultApiClient) get(reqUrl string) (*http.Response, error) {
	return http.Get(reqUrl)
}

func createTmDb(apiKey string) *TmDb {
	return &TmDb{apiKey: apiKey, client: clientFactory(), movieDetailsCache: createCache()}
}

func (tmdb *TmDb) GetConfiguration() (Configuration, error) {
	reqUrl := fmt.Sprintf("%s/configuration?api_key=%s", baseUrl, tmdb.apiKey)
	config := Configuration{}
	_, err := tmdb.sendRequest(reqUrl, &config)
	return config, err
}

func (tmdb *TmDb) SearchMovies(query string, page int) (MovieSearchResult, error) {
	reqUrl := fmt.Sprintf("%s/search/movie?api_key=%s&query=%s&page=%d", baseUrl, tmdb.apiKey, url.QueryEscape(query), page)
	result := MovieSearchResult{}
	_, err := tmdb.sendRequest(reqUrl, &result)
	return result, err
}

func (tmdb *TmDb) GetMovie(id int) (MovieDetails, error) {
	res, err := tmdb.movieDetailsCache.getOrLoad(id, func(k key) (interface{}, error) {
		reqUrl := fmt.Sprintf("%s/movie/%d?api_key=%s", baseUrl, id, tmdb.apiKey)
		mov := MovieDetails{}
		_, err := tmdb.sendRequest(reqUrl, &mov)
		log.Printf("Retrieve movie details from TMDb, movie id: %d, error: %v\n", id, err)
		return mov, err
	})
	return res.(MovieDetails), err
}

func (tmdb *TmDb) sendRequest(reqUrl string, payload interface{}) (interface{}, error) {
	tmdb.mu.Lock()
	defer tmdb.mu.Unlock()
	resp, err := tmdb.client.get(reqUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
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
	return nil, errors.New(fmt.Sprintf("error: %s, code: %d", errorStatus.Message, errorStatus.Code))
}
