package catalog

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

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

type idGenerator struct {
	v int
}

var catalogFile string

func init() {
	catalogFile = filepath.Join(config.ConfDir(), "catalog.json")
	catalogFactory = createJsonCatalog
}

func createJsonCatalog(conf *config.Config) (Catalog, error) {
	catalog := &JsonCatalog{conf: conf}
	err := catalog.Load()
	if err != nil {
		return nil, err
	}
	return catalog, nil
}

func createIdGenerator(v int) *idGenerator {
	return &idGenerator{v: v}
}

func (g *idGenerator) next() int {
	g.v++
	return g.v
}

func (ctl *JsonCatalog) All() []api.Movie {
	ctl.mu.RLock()
	defer ctl.mu.RUnlock()
	all := ctl.movies
	result := make([]api.Movie, 0, len(all))
	for _, p := range all {
		m := *p
		exists, err := file.Exists(m.Path)
		m.Available = exists && err == nil
		result = append(result, m)
	}
	return result
}

func (ctl *JsonCatalog) Find(title string) []api.Movie {
	ctl.mu.RLock()
	defer ctl.mu.RUnlock()
	ids := ctl.index.Find(title)
	result := make([]api.Movie, 0, len(ids))
	for _, id := range ids {
		p := ctl.movies[id]
		m := *p
		exists, err := file.Exists(m.Path)
		m.Available = exists && err == nil
		result = append(result, m)
	}
	return result
}

func (ctl *JsonCatalog) Get(id int) (api.Movie, bool) {
	ctl.mu.RLock()
	defer ctl.mu.RUnlock()
	m := ctl.movies[id]
	if m != nil {
		return *m, true
	}
	return api.Movie{}, false
}

func (ctl *JsonCatalog) Load() error {
	ctl.mu.Lock()
	defer ctl.mu.Unlock()
	movies, err := readCatalog(catalogFile)
	if err != nil {
		return err
	}
	err = updateCatalog(movies, ctl.conf.Dirs, ctl.conf.VideoFileExts)
	if err != nil {
		return err
	}
	index, err := CreateIndex(ctl.conf)
	if err != nil {
		return err
	}
	for _, m := range movies {
		index.Add(*m)
	}
	ctl.movies = movies
	ctl.index = index
	return nil
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
	f, err := os.OpenFile(catalogFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer func() {
		clsErr := f.Close()
		if err == nil {
			err = clsErr
		}
	}()
	encoder := json.NewEncoder(f)
	err = encoder.Encode(ctl.movies)
	return
}

func (ctl *JsonCatalog) Update(u api.Movie) (api.Movie, error) {
	ctl.mu.Lock()
	defer ctl.mu.Unlock()
	p := ctl.movies[u.Id]
	if p == nil {
		return api.Movie{}, errors.New(fmt.Sprintf("unknown movie, id: %d, title: %s", u.Id, u.Title))
	}
	// nothing else at the moment for update
	p.TMDbId = u.TMDbId

	m := *p
	exists, err := file.Exists(p.Path)
	m.Available = exists && err == nil

	ctl.save()

	return m, nil
}

func readCatalog(catalogFile string) (map[int]*api.Movie, error) {
	var err error
	var catalogExists bool
	if catalogExists, err = file.Exists(catalogFile); catalogExists && err == nil {
		f, err := os.OpenFile(catalogFile, os.O_RDONLY, 0644)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		parser := json.NewDecoder(f)
		var movieFiles map[int]*api.Movie
		if err := parser.Decode(&movieFiles); err != nil {
			return nil, err
		}
		return movieFiles, err
	}
	if err != nil {
		return nil, err
	}
	return make(map[int]*api.Movie), nil
}

func updateCatalog(files map[int]*api.Movie, dirs []string, fileExt []string) error {
	drives, err := mountedDrives()
	if err != nil {
		return err
	}
	known := make(map[string]bool, len(files))
	var maxId = 0
	for id, f := range files {
		fileDriveUnmounted := !driveMounted(drives, f)
		if exists, err := file.Exists(f.Path); err == nil && (exists || fileDriveUnmounted) {
			known[f.Path] = true
			if id > maxId {
				maxId = id
			}
		} else {
			delete(files, id)
		}
	}
	idGen := createIdGenerator(maxId)
	for _, dir := range dirs {
		if exists, err := file.Exists(dir); exists && err == nil {
			filepath.Walk(dir, func(path string, fInfo os.FileInfo, err error) error {
				if !known[path] && fInfo.Mode().IsRegular() && util.Contains(fileExt, filepath.Ext(fInfo.Name())) {
					id := idGen.next()
					title := fInfo.Name()
					drive := fileDrive(drives, path)
					driveName := ""
					if drive != nil {
						driveName = drive.name
					}
					files[id] = &api.Movie{Id: id, Path: path, Title: title, DriveName: driveName}
					log.Printf("Add file '%s' to catalog\n", path)
				}
				return nil
			})
		}
	}
	return nil
}
