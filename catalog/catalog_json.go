package catalog

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"github.com/andrew00x/gomovies/config"
	"github.com/andrew00x/gomovies/file"
	"github.com/andrew00x/gomovies/util"
)

type JsonCatalog struct {
	mu     sync.RWMutex
	movies map[int]*MovieFile
	config *config.Config
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
	catalog := &JsonCatalog{config: conf}
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

func (ctl *JsonCatalog) All() []MovieFile {
	ctl.mu.RLock()
	defer ctl.mu.RUnlock()
	all := ctl.movies
	result := make([]MovieFile, 0, len(all))
	for _, m := range all {
		result = append(result, *m)
	}
	return result
}

func (ctl *JsonCatalog) Find(title string) []MovieFile {
	ctl.mu.RLock()
	defer ctl.mu.RUnlock()
	ids := ctl.index.Find(title)
	result := make([]MovieFile, 0, len(ids))
	for _, id := range ids {
		m := ctl.movies[id]
		result = append(result, *m)
	}
	return result
}

func (ctl *JsonCatalog) Get(id int) *MovieFile {
	ctl.mu.RLock()
	defer ctl.mu.RUnlock()
	return ctl.movies[id]
}

func (ctl *JsonCatalog) Load() error {
	ctl.mu.Lock()
	defer ctl.mu.Unlock()
	movies, err := readCatalog(catalogFile)
	if err != nil {
		return err
	}
	err = updateCatalog(movies, ctl.config.Dirs, ctl.config.VideoFileExts)
	if err != nil {
		return err
	}
	index, err := CreateIndex(ctl.config)
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

func (ctl *JsonCatalog) Refresh(conf *config.Config) error {
	ctl.config = conf
	return ctl.Load()
}

func (ctl *JsonCatalog) Save() (err error) {
	ctl.mu.Lock()
	defer ctl.mu.Unlock()
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

func readCatalog(catalogFile string) (map[int]*MovieFile, error) {
	var err error
	var catalogExists bool
	if catalogExists, err = file.Exists(catalogFile); catalogExists && err == nil {
		catalog, err := os.OpenFile(catalogFile, os.O_RDONLY, 0644)
		if err != nil {
			return nil, err
		}
		defer catalog.Close()
		parser := json.NewDecoder(catalog)
		var movieFiles map[int]*MovieFile
		if err := parser.Decode(&movieFiles); err != nil {
			return nil, err
		}
		return movieFiles, err
	}
	if err != nil {
		return nil, err
	}
	return make(map[int]*MovieFile), nil
}

func updateCatalog(files map[int]*MovieFile, dirs []string, fileExt []string) error {
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
					files[id] = &MovieFile{Id: id, Path: path, Title: title, DriveName: driveName}
					log.Printf("Add file '%s' to catalog\n", path)
				}
				return nil
			})
		}
	}
	return nil
}
