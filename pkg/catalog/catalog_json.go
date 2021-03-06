package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/andrew00x/gomovies/pkg/api"
	"github.com/andrew00x/gomovies/pkg/config"
	"github.com/andrew00x/gomovies/pkg/file"
	"github.com/andrew00x/gomovies/pkg/util"
)

type JsonCatalog struct {
	mu     sync.RWMutex
	movies map[int]*api.Movie
	conf   *config.Config
	index  Index
}

var catalogFile string

func init() {
	catalogFile = filepath.Join(config.ConfDir(), "catalog.json")
	catalogFactory = createJsonCatalog
}

func createJsonCatalog(conf *config.Config) (ctl Catalog, err error) {
	ctl = &JsonCatalog{conf: conf}
	err = ctl.Load()
	return
}

func (ctl *JsonCatalog) All() []api.Movie {
	ctl.mu.RLock()
	defer ctl.mu.RUnlock()
	all := ctl.movies
	result := make([]api.Movie, 0, len(all))
	var exists bool
	var err error
	for _, p := range all {
		m := *p
		if exists, err = file.Exists(m.File); err != nil {
			log.WithFields(log.Fields{
				"err":  err,
				"file": m.File,
			}).Warn("Error occurred while trying access movie file")
		}
		m.Available = exists && err == nil
		result = append(result, m)
	}
	return result
}

func (ctl *JsonCatalog) Find(tag string) []api.Movie {
	ctl.mu.RLock()
	defer ctl.mu.RUnlock()
	ids := ctl.index.Find(tag)
	result := make([]api.Movie, 0, len(ids))
	var exists bool
	var err error
	for _, id := range ids {
		p := ctl.movies[id]
		m := *p
		if exists, err = file.Exists(m.File); err != nil {
			log.WithFields(log.Fields{"err":  err, "file": m.File}).Warn("Error occurred while trying access movie file")
		}
		m.Available = exists && err == nil
		result = append(result, m)
	}
	return result
}

func (ctl *JsonCatalog) Get(id int) (mov api.Movie, found bool) {
	ctl.mu.RLock()
	defer ctl.mu.RUnlock()
	m, ok := ctl.movies[id]
	if ok {
		found = true
		mov = *m
		if exists, err := file.Exists(m.File); err != nil {
			log.WithFields(log.Fields{"err":  err, "file": m.File}).Warn("Error occurred while trying access movie file")
		} else {
			m.Available = exists
		}
	}
	return
}

func (ctl *JsonCatalog) Load() (err error) {
	var movies map[int]*api.Movie
	ctl.mu.Lock()
	defer ctl.mu.Unlock()
	if movies, err = ctl.readCatalog(); err != nil {
		return
	}
	if err = ctl.scanForMovies(movies); err != nil {
		return
	}
	var index Index
	if index, err = CreateIndex(ctl.conf); err != nil {
		return
	}
	for _, m := range movies {
		index.Add(m.Title, m.Id)
	}
	ctl.movies = movies
	ctl.index = index
	return
}

func (ctl *JsonCatalog) Refresh() error {
	return ctl.Load()
}

func (ctl *JsonCatalog) Save() (err error) {
	ctl.mu.Lock()
	defer ctl.mu.Unlock()
	return ctl.save()
}

func (ctl *JsonCatalog) save() (err error) {
	var f *os.File
	if f, err = os.OpenFile(catalogFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
		return
	}
	defer func() {
		if clsErr := f.Close(); clsErr != nil {
			err = clsErr
		}
	}()
	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(ctl.movies)
	return
}

func (ctl *JsonCatalog) Update(u api.Movie) (m api.Movie, err error) {
	ctl.mu.Lock()
	defer ctl.mu.Unlock()
	p := ctl.movies[u.Id]
	if p == nil {
		err = fmt.Errorf("unknown movie, id: %d, title: %s", u.Id, u.Title)
		return
	}
	exists := false
	if exists, err = file.Exists(p.File); err != nil {
		return
	}
	// nothing else at the moment for update
	p.TMDbId = u.TMDbId
	p.Available = exists
	m = *p

	err = ctl.save()

	return
}

func (ctl *JsonCatalog) AddTag(tag string, id int) error {
	ctl.mu.Lock()
	defer ctl.mu.Unlock()
	m := ctl.movies[id]
	if m == nil {
		return fmt.Errorf("unable add tag for unknown movie, id: %d", id)
	}
	ctl.index.Add(tag, id)
	log.WithFields(log.Fields{"file": m.File, "tag": tag}).Info("Add tag for movie")
	return nil
}

func (ctl *JsonCatalog) readCatalog() (movies map[int]*api.Movie, err error) {
	movies = make(map[int]*api.Movie)
	var exists bool
	if exists, err = file.Exists(catalogFile); exists && err == nil {
		var f *os.File
		if f, err = os.OpenFile(catalogFile, os.O_RDONLY, 0644); err != nil {
			return
		}
		defer func() {
			if clsErr := f.Close(); clsErr != nil {
				err = clsErr
			}
		}()
		parser := json.NewDecoder(f)
		err = parser.Decode(&movies)
	}
	return
}

func (ctl *JsonCatalog) scanForMovies(files map[int]*api.Movie) (err error) {
	var drives []*drive
	if drives, err = mountedDrives(); err != nil {
		return
	}
	known := make(map[string]bool, len(files))
	var maxID = 0
	for id, f := range files {
		fileDriveMounted := driveMounted(drives, f)
		exists := false
		if exists, err = file.Exists(f.File); err == nil && (exists || !fileDriveMounted) {
			known[f.File] = true
			if id > maxID {
				maxID = id
			}
		} else if err != nil {
			return
		} else {
			delete(files, id)
		}
	}
	idGen := util.CreateIdGenerator(maxID)
	for _, dir := range ctl.conf.Dirs {
		exists := false
		if exists, err = file.Exists(dir); exists && err == nil {
			err = filepath.Walk(dir, func(path string, fInfo os.FileInfo, _ error) error {
				if !known[path] && fInfo.Mode().IsRegular() && util.Contains(ctl.conf.VideoFileExts, filepath.Ext(fInfo.Name())) {
					id := idGen.Next()
					title := fInfo.Name()
					drive := fileDrive(drives, path)
					driveName := ""
					if drive != nil {
						driveName = drive.name
					}
					files[id] = &api.Movie{Id: id, File: path, Title: title, DriveName: driveName}
					log.WithFields(log.Fields{"file": path}).Debug("Add file to catalog")
				}
				return nil
			})
		}
		if err != nil {
			return
		}
	}
	return
}
