package details

import (
	"encoding/json"
	"fmt"
	"github.com/andrew00x/gomovies/pkg/api"
	"github.com/andrew00x/gomovies/pkg/file"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

func GetDetails(path, lang string) (mov api.MovieDetails, err error) {
	detailsFile := fmt.Sprintf("%s.%s.%s", path, lang, "json")
	exists := false
	if exists, err = file.Exists(detailsFile); exists && err == nil {
		mov, err = loadDetails(detailsFile)
		if err == nil {
			log.WithFields(log.Fields{"movie_path": path, "title": mov.OriginalTitle, "lang": lang, "file": detailsFile}).Info("Found details in local file")
		}
	} else if err == nil {
		detailsFile = filepath.Join(filepath.Dir(path), fmt.Sprintf("gomovies-details.%s.json", lang))
		if exists, err = file.Exists(detailsFile); exists && err == nil {
			mov, err = loadDetails(detailsFile)
			if err == nil {
				log.WithFields(log.Fields{"movie_path": path, "title": mov.OriginalTitle, "lang": lang, "file": detailsFile}).Info("Found details in local file")
			}
		}
	}
	if err != nil {
		log.WithFields(log.Fields{"movie_path": path, "lang": lang, "err": err}).Error("Error occurred while retrieving movie details from TMDb")
	} else if !exists {
		err = fmt.Errorf("there is no details for movie %s, lang %s", path, lang)
	}
	return
}

func loadDetails(filename string) (mov api.MovieDetails, err error) {
	var f *os.File
	if f, err = os.OpenFile(filename, os.O_RDONLY, 0644); err != nil {
		return
	}
	defer func() {
		if clsErr := f.Close(); clsErr != nil {
			err = clsErr
		}
	}()
	parser := json.NewDecoder(f)
	err = parser.Decode(&mov)
	return
}
