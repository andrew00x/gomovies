package details

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/andrew00x/gomovies/pkg/api"
	"github.com/stretchr/testify/assert"
)

var moviesDir string

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func setup() {
	tmp := os.Getenv("TMPDIR")
	moviesDir = filepath.Join(tmp, "DetailsTest")
	err := os.RemoveAll(moviesDir)
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}
}

func TestGetDetailFromDedicatedFile(t *testing.T) {
	mustCreateDir(filepath.Join(moviesDir, "back to the future"))
	movieDetails := api.MovieDetails{OriginalTitle: "Back to the Future", Overview: "Marty McFly is accidentally sent back in time..."}
	movieDetailsFile := filepath.Join(moviesDir, "back to the future", "back to the future 1.avi.en.json")
	mustSaveMovieDetailsFile(movieDetailsFile, movieDetails)
	defer func() { mustRemoveFiles(movieDetailsFile) }()

	loadedMovieDetails, err := GetDetails(filepath.Join(moviesDir, "back to the future", "back to the future 1.avi"), "en")
	assert.Nil(t, err)
	assert.NotNil(t, loadedMovieDetails)
	assert.Equal(t, movieDetails, loadedMovieDetails)
}

func TestGetDetailFromGlobalFile(t *testing.T) {
	mustCreateDir(filepath.Join(moviesDir, "back to the future"))
	movieDetails := api.MovieDetails{OriginalTitle: "Back to the Future", Overview: "Marty McFly is accidentally sent back in time...."}
	movieDetailsFile := filepath.Join(moviesDir, "back to the future", "gomovies-details.en.json")
	mustSaveMovieDetailsFile(movieDetailsFile, movieDetails)
	defer func() { mustRemoveFiles(movieDetailsFile) }()

	loadedMovieDetails, err := GetDetails(filepath.Join(moviesDir, "back to the future", "back to the future 1.avi"), "en")
	assert.Nil(t, err)
	assert.NotNil(t, loadedMovieDetails)
	assert.Equal(t, movieDetails, loadedMovieDetails)
}

func TestReturnsErrorWhenNeitherDedicatedFileNorGlobalFileExists(t *testing.T) {
	mustCreateDir(filepath.Join(moviesDir, "back to the future"))

	movie := filepath.Join(moviesDir, "back to the future", "back to the future 1.avi")
	_, err := GetDetails(movie, "en")
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("there is no details for movie %s, lang %s", movie, "en"), err.Error())
}

func mustSaveMovieDetailsFile(path string, movie api.MovieDetails) {
	var err error
	var file *os.File
	if file, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644); err != nil {
		log.Fatal(err)
	}
	encoder := json.NewEncoder(file)
	if err = encoder.Encode(movie); err != nil {
		log.Fatal(err)
	}
	if err = file.Close(); err != nil {
		log.Fatal(err)
	}
}

func mustCreateDir(dir string) {
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		log.Fatal(err)
	}
}

func mustRemoveFiles(paths ...string) {
	for _, path := range paths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			log.Fatal(err)
		}
	}
}
