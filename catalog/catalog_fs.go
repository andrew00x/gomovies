package catalog

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"strings"
	"github.com/andrew00x/gomovies/config"
	"github.com/andrew00x/gomovies/file"
	"github.com/andrew00x/gomovies/util"
)

type MovieFile struct {
	Id        int    `json:"id"`
	Title     string `json:"title"`
	Path      string `json:"path"`
	DriveName string `json:"drive"`
}

type FSCatalog struct {
	refreshLock sync.Mutex
	movies      map[int]*MovieFile
	catalogFile string
	index       *index
}

type idGenerator struct {
	v  int
}

type drive struct {
	devSpec    string
	mountPoint string
	name       string
}

var devcd string
var etccd string

func init() {
	devcd = "/dev"
	etccd = "/etc"
	factory = func(conf *config.Config) (Catalog, error) {
		var err error
		files, err := readCatalog(conf.CatalogFile)
		if err != nil {
			return nil, err
		}
		err = updateCatalog(files, conf.Dirs, conf.VideoFileExts)
		if err != nil {
			return nil, err
		}

		idx := newIndex(len(files))
		for _, m := range files {
			idx.add(*m)
		}
		catalog := &FSCatalog{movies: files, catalogFile: conf.CatalogFile, index: idx}
		return catalog, nil
	}
}

func newIdGenerator(v int) *idGenerator {
	return &idGenerator{v: v}
}

func (g *idGenerator) next() int {
	g.v++
	return g.v
}

func (ctl *FSCatalog) Get(id int) *MovieFile {
	return ctl.movies[id]
}

func (ctl *FSCatalog) Find(title string) []MovieFile {
	ids := ctl.index.find(title)
	result := make([]MovieFile, 0, len(ids))
	for _, id := range ids {
		m := ctl.movies[id]
		result = append(result, *m)
	}
	return result
}

func (ctl *FSCatalog) All() []MovieFile {
	all := ctl.movies
	result := make([]MovieFile, 0, len(all))
	for _, m := range all {
		result = append(result, *m)
	}
	return result
}

func (ctl *FSCatalog) Save() (err error) {
	f, err := os.OpenFile(ctl.catalogFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
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

func (ctl *FSCatalog) Refresh(conf *config.Config) error {
	ctl.refreshLock.Lock()
	defer ctl.refreshLock.Unlock()
	err := updateCatalog(ctl.movies, conf.Dirs, conf.VideoFileExts)
	if err == nil {
		idx := newIndex(len(ctl.movies))
		for _, m := range ctl.movies {
			idx.add(*m)
		}
		ctl.index = idx
		ctl.catalogFile = conf.CatalogFile
	}

	return err
}

func readCatalog(f string) (map[int]*MovieFile, error) {
	var err error
	var catalogExists bool
	if catalogExists, err = file.Exists(f); catalogExists && err == nil {
		catalog, err := os.OpenFile(f, os.O_RDONLY, 0644)
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
	idGen := newIdGenerator(maxId)
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

func fileDrive(drives []*drive, file string) *drive {
	return findDrive(drives, func(d *drive) bool { return strings.HasPrefix(file, d.mountPoint) })
}

func driveMounted(drives []*drive, f *MovieFile) bool {
	return findDrive(drives, func(d *drive) bool { return d.name == f.DriveName }) != nil
}

func findDrive(drives []*drive, predicate func(*drive) bool) *drive {
	for _, drive := range drives {
		if predicate(drive) {
			return drive
		}
	}
	return nil
}

func mountedDrives() ([]*drive, error) {
	drives := make(map[string]*drive)

	err := drivesByLabel(drives)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	err = drivesById(drives)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	err = setMountPoints(drives)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	mounted := make([]*drive, 0, len(drives))
	for _, drive := range drives {
		if drive.mountPoint != "" {
			mounted = append(mounted, drive)
		}
	}

	return mounted, nil
}

func drivesByLabel(drives map[string]*drive) error {
	byLabelDir := filepath.Join(devcd, "disk", "by-label")
	disksByLabel, err := ioutil.ReadDir(byLabelDir)
	if err != nil {
		return err
	}
	for _, link := range disksByLabel {
		label := link.Name()
		devFile, err := os.Readlink(filepath.Join(byLabelDir, label))
		if err != nil {
			return err
		}
		devName := filepath.Base(devFile)
		drives[devName] = &drive{
			devSpec: filepath.Join(devcd, devName),
			name:    label}
	}
	return nil
}

func drivesById(drives map[string]*drive) error {
	byIdDir := filepath.Join(devcd, "disk", "by-id")

	disksById, err := ioutil.ReadDir(byIdDir)
	if err != nil {
		return err
	}
	for _, link := range disksById {
		id := link.Name()
		devFile, err := os.Readlink(filepath.Join(byIdDir, id))
		if err != nil {
			return err
		}
		devName := filepath.Base(devFile)
		_, ok := drives[devName]
		if !ok {
			drives[devName] = &drive{
				devSpec: filepath.Join(devcd, devName),
				name:    id}
		}
	}
	return nil
}

func setMountPoints(drives map[string]*drive) error {
	mtab := filepath.Join(etccd, "mtab")
	return file.ReadLines(mtab, func(line string) bool {
		fields := strings.Fields(line)
		devName := filepath.Base(fields[0])
		drive, ok := drives[devName]
		if ok {
			drive.mountPoint = fields[1]
		}
		return true
	})
}
