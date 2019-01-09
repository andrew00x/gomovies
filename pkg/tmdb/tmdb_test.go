package tmdb

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"github.com/stretchr/testify/assert"
)

const fakeApiKey = "123"

type mockApiClient struct{}

var doGetFunc func(reqUrl string) (*http.Response, error)

func (*mockApiClient) get(reqUrl string) (*http.Response, error) {
	if doGetFunc != nil {
		return doGetFunc(reqUrl)
	}
	return &http.Response{}, nil
}

func TestMain(m *testing.M) {
	clientFactory = func() apiClient { return &mockApiClient{} }
	code := m.Run()
	os.Exit(code)
}

func TestGetConfiguration(t *testing.T) {
	expectedReqUrl := fmt.Sprintf("%s/configuration?api_key=%s", baseUrl, fakeApiKey)
	doGetFunc = func(reqUrl string) (*http.Response, error) {
		if reqUrl == expectedReqUrl {
			resp := http.Response{StatusCode: 200, Body: ioutil.NopCloser(strings.NewReader(configurationResponseBody))}
			return &resp, nil
		}
		return nil, errors.New(fmt.Sprintf("invalid request url: %s, expected to be %s", reqUrl, expectedReqUrl))
	}
	tmdb := GetTmDbInstance(fakeApiKey)
	result, err := tmdb.GetConfiguration()
	if err != nil {
		t.Fatal(err)
	}
	expectedResult := Configuration{
		Images{
			BaseUrl:       "http://image.tmdb.org/t/p/",
			SecureBaseUrl: "https://image.tmdb.org/t/p/",
			BackdropSizes: []string{"b1", "b2", "b3", "original"},
			LogoSizes:     []string{"l1", "l2", "l3", "original"},
			PosterSizes:   []string{"p1", "p2", "p3", "original"},
			ProfileSizes:  []string{"f1", "f2", "f3", "original"},
			StillSizes:    []string{"s1", "s2", "s3", "original"},
		}}
	assert.Equal(t, expectedResult, result)
}

func TestSearchMovie(t *testing.T) {
	expectedReqUrl := fmt.Sprintf("%s/search/movie?api_key=%s&query=%s&page=1", baseUrl, fakeApiKey, url.QueryEscape("brave heart"))
	doGetFunc = func(reqUrl string) (*http.Response, error) {
		if reqUrl == expectedReqUrl {
			resp := http.Response{StatusCode: 200, Body: ioutil.NopCloser(strings.NewReader(searchMoviesResponseBody))}
			return &resp, nil
		}
		return nil, errors.New(fmt.Sprintf("invalid request url: %s, expected to be %s", reqUrl, expectedReqUrl))
	}
	tmdb := GetTmDbInstance(fakeApiKey)
	result, err := tmdb.SearchMovies("brave heart")
	if err != nil {
		t.Fatal(err)
	}
	expectedResult := []MovieShort{{
		Id:               1,
		Title:            "Braveheart",
		OriginalTitle:    "Braveheart",
		OriginalLanguage: "en",
		Overview:         "the best",
		ReleaseDate:      "1995-05-01",
		PosterPath:       "/braveheart_poster.jpg",
		BackdropPath:     "/braveheart_backdrop.jpg",
	}}
	assert.Equal(t, expectedResult, result)
}

func TestGetMovie(t *testing.T) {
	expectedReqUrl := fmt.Sprintf("%s/movie/%d?api_key=%s", baseUrl, 123, fakeApiKey)
	doGetFunc = func(reqUrl string) (*http.Response, error) {
		if reqUrl == expectedReqUrl {
			resp := http.Response{StatusCode: 200, Body: ioutil.NopCloser(strings.NewReader(movieDetailsResponse))}
			return &resp, nil
		}
		return nil, errors.New(fmt.Sprintf("invalid request url: %s, expected to be %s", reqUrl, expectedReqUrl))
	}
	tmdb := GetTmDbInstance(fakeApiKey)
	result, err := tmdb.GetMovie(123)
	if err != nil {
		t.Fatal(err)
	}
	expectedResult := MovieDetails{
		Id:               1,
		Title:            "Braveheart",
		OriginalTitle:    "Braveheart",
		OriginalLanguage: "en",
		Overview:         "the best",
		ReleaseDate:      "1995-05-01",
		PosterPath:       "/braveheart_poster.jpg",
		BackdropPath:     "/braveheart_backdrop.jpg",
		ImdbId:           "imdb2",
		Budget:           10000000,
		Revenue:          100000000,
		TagLine:          "Every man dies. Not every man truly lives.",
		Genres: []Genre{
			{Id: 1, Name: "Action"},
			{Id: 2, Name: "Drama"},
			{Id: 3, Name: "History"},
			{Id: 4, Name: "War"},
		},
		ProductionCompanies: []Company{
			{Id: 1234, LogoPath: "icon_entertainment.jpg", Name: "Icon Entertainment International"},
		},
		ProductionCountries: []Country{
			{Code: "US", Name: "USA"},
		},
	}
	assert.Equal(t, expectedResult, result)
}

func TestHandleApiErrorStatus(t *testing.T) {
	doGetFunc = func(reqUrl string) (*http.Response, error) {
		return &http.Response{StatusCode: 400, Body: ioutil.NopCloser(strings.NewReader(apiErrorResponse))}, nil
	}
	tmdb := GetTmDbInstance(fakeApiKey)
	_, err := tmdb.GetMovie(0)
	if err == nil {
		t.Fatal("error expected")
	}
	expectedErrorMessage := fmt.Sprintf("error: %s, code: %d", "Invalid movie id", 1)
	assert.Equal(t, expectedErrorMessage, err.Error())
}

const configurationResponseBody = `
{
"images": {
    "base_url": "http://image.tmdb.org/t/p/",
    "secure_base_url": "https://image.tmdb.org/t/p/",
    "backdrop_sizes": [
      "b1",
      "b2",
      "b3",
      "original"
    ],
    "logo_sizes": [
      "l1",
      "l2",
      "l3",
      "original"
    ],
    "poster_sizes": [
      "p1",
      "p2",
      "p3",
      "original"
    ],
    "profile_sizes": [
      "f1",
      "f2",
      "f3",
      "original"
    ],
    "still_sizes": [
      "s1",
      "s2",
      "s3",
      "original"
    ]
  }
}`

const searchMoviesResponseBody = `
{
  "page": 1,
  "total_results": 1,
  "total_pages": 1,
  "results": [
    {
      "id": 1,
      "title": "Braveheart",
      "poster_path": "\/braveheart_poster.jpg",
      "original_language": "en",
      "original_title": "Braveheart",
      "backdrop_path": "\/braveheart_backdrop.jpg",
      "overview": "the best",
      "release_date": "1995-05-01"
    }
  ]
}`

const movieDetailsResponse = `
{
  "id": 1,
  "title": "Braveheart",
  "imdb_id": "imdb2",
  "original_language": "en",
  "original_title": "Braveheart",
  "poster_path": "\/braveheart_poster.jpg",
  "backdrop_path": "\/braveheart_backdrop.jpg",
  "overview": "the best",
  "release_date": "1995-05-01",
  "budget":  10000000,
  "revenue": 100000000,
  "tagline": "Every man dies. Not every man truly lives.",
  "genres": [
    {"id": 1, "name": "Action"},
    {"id": 2, "name": "Drama"},
    {"id": 3, "name": "History"},
    {"id": 4, "name": "War"}
  ],
  "production_companies": [
    {
      "id": 1234,
      "logo_path": "icon_entertainment.jpg",
      "name": "Icon Entertainment International"
    }
  ],
  "production_countries": [
    {
      "iso_3166_1": "US",
      "name": "USA"
    }
  ]
}`

const apiErrorResponse = `
{
  "status_code": 1,
  "status_message": "Invalid movie id"
}`
