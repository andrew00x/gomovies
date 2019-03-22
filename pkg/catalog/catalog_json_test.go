package catalog

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"text/template"

	"github.com/andrew00x/gomovies/pkg/api"
	"github.com/andrew00x/gomovies/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func TestLoadCatalog(t *testing.T) {
	mustRemoveFiles(catalogFile)
	mustCreateDir(filepath.Join(moviesDir, "back to the future"))
	files := []string{
		filepath.Join(moviesDir, "brave heart.mkv"),
		filepath.Join(moviesDir, "back to the future", "back to the future 1.avi"),
	}
	defer func() { mustRemoveFiles(files...) }()

	movies := make([]api.Movie, 0, len(files))
	id := 1
	for _, f := range files {
		mustCreateFile(f)
		movies = append(movies, api.Movie{Id: id, Path: f, Title: filepath.Base(f), DriveName: sda1})
		id++
	}

	mustSaveCatalogFile(movies)

	index := indexMock{[]api.Movie{}, []int{}}
	indexFactory = func(_ *config.Config) (Index, error) { return &index, nil }

	catalog, err := createJsonCatalog(&config.Config{})
	assert.Nil(t, err)

	result := catalog.All()
	sort.Sort(byId(result))
	expected := make([]api.Movie, 0, len(movies))
	for _, m := range movies {
		m.Available = true
		expected = append(expected, m)
	}
	assert.Equal(t, expected, result)

	indexed := index.added
	sort.Sort(byId(indexed))
	assert.Equal(t, movies, indexed)
}

func TestCreateNewCatalog(t *testing.T) {
	mustRemoveFiles(catalogFile)
	mustCreateDir(filepath.Join(moviesDir, "start wars"))
	files := []string{
		filepath.Join(moviesDir, "gladiator.mkv"),
		filepath.Join(moviesDir, "start wars", "start wars 1.avi"),
		filepath.Join(moviesDir, "start wars", "start wars 2.mkv"),
	}
	defer func() { mustRemoveFiles(files...) }()

	movies := make([]api.Movie, 0, len(files))
	id := 1
	for _, f := range files {
		mustCreateFile(f)
		movies = append(movies, api.Movie{Id: id, Path: f, Title: filepath.Base(f), DriveName: sda1})
		id++
	}

	index := indexMock{[]api.Movie{}, []int{}}
	indexFactory = func(_ *config.Config) (Index, error) { return &index, nil }

	conf := &config.Config{Dirs: []string{moviesDir}, VideoFileExts: []string{".mkv", ".avi"}}
	catalog, err := createJsonCatalog(conf)
	assert.Nil(t, err)

	expected := make([]api.Movie, 0, len(movies))
	for _, m := range movies {
		m.Available = true
		expected = append(expected, m)
	}
	result := catalog.All()
	sort.Sort(byId(result))
	assert.Equal(t, expected, result)

	indexed := index.added
	sort.Sort(byId(indexed))
	assert.Equal(t, movies, indexed)
}

func TestCreateAndUpdateCatalog(t *testing.T) {
	mustRemoveFiles(catalogFile)
	mustCreateDir(filepath.Join(moviesDir, "lethal weapon"))
	files := []string{
		filepath.Join(moviesDir, "green mile.mkv"),
		filepath.Join(moviesDir, "lethal weapon", "lethal weapon 1.avi"),
	}
	defer func() { mustRemoveFiles(files...) }()

	movies := make([]api.Movie, 0, len(files))
	id := 1
	for _, f := range files {
		mustCreateFile(f)
		movies = append(movies, api.Movie{Id: id, Path: f, Title: filepath.Base(f), DriveName: sda1})
		id++
	}

	mustSaveCatalogFile(movies)

	newFile := filepath.Join(moviesDir, "lethal weapon", "lethal weapon 4.mkv")
	mustCreateFile(newFile)
	files = append(files, newFile)
	movies = append(movies, api.Movie{Id: id, Path: newFile, Title: filepath.Base(newFile), DriveName: sda1})

	index := indexMock{[]api.Movie{}, []int{}}
	indexFactory = func(_ *config.Config) (Index, error) { return &index, nil }

	conf := &config.Config{Dirs: []string{moviesDir}, VideoFileExts: []string{".mkv", ".avi"}}
	catalog, err := createJsonCatalog(conf)
	assert.Nil(t, err)

	expected := make([]api.Movie, 0, len(movies))
	for _, m := range movies {
		m.Available = true
		expected = append(expected, m)
	}
	result := catalog.All()
	sort.Sort(byId(result))
	assert.Equal(t, expected, result)

	indexed := index.added
	sort.Sort(byId(indexed))
	assert.Equal(t, movies, indexed)
}

func TestSaveCatalog(t *testing.T) {
	mustRemoveFiles(catalogFile)
	files := []string{
		filepath.Join(moviesDir, "shawshank redemption.mkv"),
		filepath.Join(moviesDir, "fight club.avi"),
	}

	movies := make(map[int]*api.Movie)
	id := 1
	for _, f := range files {
		movies[id] = &api.Movie{Id: id, Title: filepath.Base(f), Path: f, DriveName: sda1}
		id++
	}
	catalog := &JsonCatalog{movies: movies}
	err := catalog.Save()
	assert.Nil(t, err)

	f, err := os.Open(catalogFile)
	assert.Nil(t, err)

	defer func() {
		clsErr := f.Close()
		assert.Nil(t, clsErr)
	}()
	parser := json.NewDecoder(f)
	var saved map[int]*api.Movie
	err = parser.Decode(&saved)
	assert.Nil(t, err)

	assert.Equal(t, movies, saved)
}

func TestGetById(t *testing.T) {
	files := []string{
		filepath.Join(moviesDir, "hobbit 1 (an unexpected journey).mkv"),
		filepath.Join(moviesDir, "hobbit 2 (the desolation of smaug).mkv"),
	}
	var id = 1
	movies := make(map[int]*api.Movie)
	for _, f := range files {
		movies[id] = &api.Movie{Id: id, Path: f, Title: filepath.Base(f), DriveName: sda1}
		id++
	}
	catalog := &JsonCatalog{movies: movies}

	result, ok := catalog.Get(1)

	assert.True(t, ok)
	assert.Equal(t, *movies[1], result)
}

func TestGetByIdReturnsEmptyResultWhenFileDoesNotExist(t *testing.T) {
	movies := make(map[int]*api.Movie)
	catalog := &JsonCatalog{movies: movies}

	result, ok := catalog.Get(1)

	assert.False(t, ok)
	assert.Equal(t, api.Movie{}, result)
}

func TestFindByNameInCatalog(t *testing.T) {
	mustRemoveFiles(catalogFile)
	files := []string{
		filepath.Join(moviesDir, "hobbit 1 (an unexpected journey).mkv"),
		filepath.Join(moviesDir, "hobbit 2 (the desolation of smaug).mkv"),
		filepath.Join(moviesDir, "hobbit 3 (battle of five armies).mkv"),
	}
	var id = 1
	movies := make(map[int]*api.Movie)
	for _, f := range files {
		movies[id] = &api.Movie{Id: id, Path: f, Title: filepath.Base(f), DriveName: sda1}
		id++
	}

	index := indexMock{added: []api.Movie{}, found: []int{1, 3}}
	catalog := &JsonCatalog{movies: movies, index: &index}

	result := catalog.Find("whatever we have in index")

	expected := []api.Movie{*movies[1], *movies[3]}
	assert.Equal(t, expected, result)
}

func TestRemoveNonexistentFilesFromCatalog(t *testing.T) {
	mustRemoveFiles(catalogFile)
	files := []string{
		filepath.Join(moviesDir, "home alone 1.avi"),
		filepath.Join(moviesDir, "home alone 2.avi"),
	}
	mustCreateFile(files[0])
	defer mustRemoveFiles(files[0])
	var id = 1
	movies := make([]api.Movie, 0, len(files))
	for _, f := range files {
		movies = append(movies, api.Movie{Id: id, Path: f, Title: filepath.Base(f), DriveName: sda1})
		id++
	}
	mustSaveCatalogFile(movies)

	index := indexMock{[]api.Movie{}, []int{}}
	indexFactory = func(_ *config.Config) (Index, error) { return &index, nil }

	conf := &config.Config{Dirs: []string{moviesDir}, VideoFileExts: []string{".avi"}}
	catalog, err := createJsonCatalog(conf)
	assert.Nil(t, err)

	m := movies[0]
	m.Available = true
	expected := []api.Movie{m}

	result := catalog.All()
	sort.Sort(byId(result))
	assert.Equal(t, expected, result)

	indexed := index.added
	sort.Sort(byId(indexed))
	assert.Equal(t, []api.Movie{movies[0]}, indexed)
}

func TestKeepNonexistentFilesWhenCorrespondedDriveIsUnmounted(t *testing.T) {
	mustRemoveFiles(catalogFile)
	moviesDir2 := filepath.Join(testRoot, "movies2")
	mustCreateDir(moviesDir2)
	files := []string{
		filepath.Join(moviesDir, "rush hour 1.avi"),
		filepath.Join(moviesDir, "rush hour 2.mkv"),
	}
	var id = 1
	movies := make([]api.Movie, 0, len(files))
	for _, f := range files {
		mustCreateFile(f)
		movies = append(movies, api.Movie{Id: id, Path: f, Title: filepath.Base(f), DriveName: sda1})
		id++
	}
	defer func() { mustRemoveFiles(files...) }()
	nonexistent := filepath.Join(moviesDir2, "rush hour 3.mkv")
	files = append(files, nonexistent)
	movies = append(movies, api.Movie{Id: id, Path: nonexistent, Title: filepath.Base(nonexistent), DriveName: "unmounted-drive"})
	mustSaveCatalogFile(movies)

	index := indexMock{[]api.Movie{}, []int{}}
	indexFactory = func(_ *config.Config) (Index, error) { return &index, nil }

	conf := &config.Config{Dirs: []string{testRoot}, VideoFileExts: []string{".mkv", ".avi"}}
	catalog, err := createJsonCatalog(conf)
	assert.Nil(t, err)

	expected := make([]api.Movie, 0, len(movies))
	for _, m := range movies {
		m.Available = m.Path != nonexistent
		expected = append(expected, m)
	}
	result := catalog.All()
	sort.Sort(byId(result))
	assert.Equal(t, expected, result)

	indexed := index.added
	sort.Sort(byId(indexed))
	assert.Equal(t, movies, indexed)
}

func TestRefreshCatalog(t *testing.T) {
	mustRemoveFiles(catalogFile)
	files := []string{
		filepath.Join(moviesDir, "rocky 1.avi"),
		filepath.Join(moviesDir, "rocky 2.mkv"),
		filepath.Join(moviesDir, "rocky 4.avi"),
	}
	movies := make([]api.Movie, 0, len(files))
	id := 1
	for _, f := range files {
		mustCreateFile(f)
		movies = append(movies, api.Movie{Id: id, Path: f, Title: filepath.Base(f), DriveName: sda1})
		id++
	}
	mustSaveCatalogFile(movies)
	defer func() { mustRemoveFiles(files...) }()

	index := indexMock{[]api.Movie{}, []int{}}
	indexFactory = func(_ *config.Config) (Index, error) { return &index, nil }

	conf := &config.Config{Dirs: []string{moviesDir}, VideoFileExts: []string{".mkv", ".avi"}}
	catalog, err := createJsonCatalog(conf)
	assert.Nil(t, err)

	expected := make([]api.Movie, 0, len(movies))
	for _, m := range movies {
		m.Available = true
		expected = append(expected, m)
	}
	result := catalog.All()
	sort.Sort(byId(result))
	assert.Equal(t, expected, result)

	mustRemoveFiles(files[2])
	newFile := filepath.Join(moviesDir, "rocky 5.mkv")
	files[2] = newFile
	mustCreateFile(newFile)
	movies = movies[0:2]
	movies = append(movies, api.Movie{Id: 3, Path: newFile, Title: filepath.Base(newFile), DriveName: sda1})
	index = indexMock{[]api.Movie{}, []int{}}

	err = catalog.Refresh()
	assert.Nil(t, err)

	expected = make([]api.Movie, 0, len(movies))
	for _, m := range movies {
		m.Available = true
		expected = append(expected, m)
	}
	result = catalog.All()
	sort.Sort(byId(result))
	assert.Equal(t, expected, result)

	indexed := index.added
	sort.Sort(byId(indexed))
	assert.Equal(t, movies, indexed)
}

func TestUpdateCatalog(t *testing.T) {
	mustRemoveFiles(catalogFile)
	files := []string{
		filepath.Join(moviesDir, "hobbit 1 (an unexpected journey).mkv"),
		filepath.Join(moviesDir, "hobbit 2 (the desolation of smaug).mkv"),
	}
	var id = 1
	movies := make(map[int]*api.Movie)
	for _, f := range files {
		movies[id] = &api.Movie{Id: id, Path: f, Title: filepath.Base(f), DriveName: sda1}
		id++
	}

	catalog := &JsonCatalog{movies: movies}

	updated, err := catalog.Update(api.Movie{Id: 1, TMDbId: 101})
	assert.Nil(t, err)
	assert.Equal(t, 101, movies[1].TMDbId)
	assert.Equal(t, 101, updated.TMDbId)

	f, err := os.Open(catalogFile)
	assert.Nil(t, err)

	defer func() {
		clsErr := f.Close()
		assert.Nil(t, clsErr)
	}()
	parser := json.NewDecoder(f)
	var saved map[int]*api.Movie
	err = parser.Decode(&saved)
	assert.Nil(t, err)

	assert.Equal(t, movies, saved)
}

func TestUpdateCatalogFailsWhenUpdatedFileDoesNotExist(t *testing.T) {
	movies := make(map[int]*api.Movie)
	catalog := &JsonCatalog{movies: movies}

	_, err := catalog.Update(api.Movie{Id: 1})
	assert.NotNil(t, err)
	assert.Equal(t, "unknown movie, id: 1, title: ", err.Error())
}

type byId []api.Movie

func (m byId) Less(i, j int) bool { return m[i].Id < m[j].Id }
func (m byId) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m byId) Len() int           { return len(m) }

var testRoot string
var moviesDir string

const sda = "usb-WDC_WD64_00AAKS-00A7B0_00A1234567E7-0:0"
const sda1 = sda + "-part1"

func setup() {
	tmp := os.Getenv("TMPDIR")
	testRoot = filepath.Join(tmp, "CatalogTest")
	err := os.RemoveAll(testRoot)
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}
	createDevFiles()
	etcDir = filepath.Join(testRoot, "etc")
	mustCreateDir(etcDir)
	moviesDir = filepath.Join(testRoot, "movies")
	mustCreateDir(moviesDir)
	writeMtabFile(etcDir)
	catalogFile = filepath.Join(testRoot, "catalog.json")
	indexFactory = func(_ *config.Config) (Index, error) { return &indexMock{}, nil }
}

func writeMtabFile(etc string) {
	var wd string
	var err error
	if wd, err = os.Getwd(); err != nil {
		log.Fatal(err)
	}
	var tpl *template.Template
	if tpl, err = template.New("mtab").ParseFiles(filepath.Join(wd, "testdata", "etc", "mtab")); err != nil {
		log.Fatal(err)
	}
	var mtabFile *os.File
	if mtabFile, err = os.OpenFile(filepath.Join(etc, "mtab"), os.O_RDWR|os.O_CREATE, 0644); err != nil {
		log.Fatal(err)
	}
	defer func() {
		if clsErr := mtabFile.Close(); clsErr != nil {
			log.Println(clsErr)
		}
	}()
	if err = tpl.Execute(mtabFile, moviesDir); err != nil {
		log.Fatal(err)
	}
}

func createDevFiles() {
	dev := filepath.Join(testRoot, "dev")
	disk := filepath.Join(dev, "disk")
	byId := filepath.Join(disk, "by-id")
	byLabel := filepath.Join(disk, "by-label")

	mustCreateDir(dev)
	mustCreateDir(disk)
	mustCreateDir(byId)
	mustCreateDir(byLabel)

	drives := map[string]struct {
		id    string
		label string
	}{
		"mmcblk0p0": {id: "mmc-SL16G_0x2a1994a5"},
		"mmcblk0p1": {id: "mmc-SL16G_0x2a1994a5-part1", label: "RECOVERY"},
		"mmcblk0p2": {id: "mmc-SL16G_0x2a1994a5-part2"},
		"mmcblk0p5": {id: "mmc-SL16G_0x2a1994a5-part5", label: "SETTINGS"},
		"mmcblk0p6": {id: "mmc-SL16G_0x2a1994a5-part6", label: "boot"},
		"mmcblk0p7": {id: "mmc-SL16G_0x2a1994a5-part7", label: "root"},
		"mmcblk0p8": {id: "mmc-SL16G_0x2a1994a5-part8", label: "data"},
		"sda":       {id: sda},
		"sda1":      {id: sda1}}
	var err error
	var wd string
	if wd, err = os.Getwd(); err != nil {
		log.Fatal(err)
	}
	for n, d := range drives {
		mustCreateFile(filepath.Join(dev, n))
		if err = os.Chdir(byId); err != nil {
			log.Fatal(err)
		}
		if err = os.Symlink(filepath.Join("../..", n), filepath.Join(byId, d.id)); err != nil {
			log.Fatal(err)
		}
		if d.label != "" {
			if err = os.Chdir(byId); err != nil {
				log.Fatal(err)
			}
			if err = os.Symlink(filepath.Join("../..", n), filepath.Join(byLabel, d.label)); err != nil {
				log.Fatal(err)
			}
		}
	}
	if err = os.Chdir(wd); err != nil {
		log.Fatal(err)
	}
	devcd = dev
}

func mustCreateDir(dir string) {
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		log.Fatal(err)
	}
}

func mustCreateFile(path string) {
	var f *os.File
	var err error
	if f, err = os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0644); err != nil {
		log.Fatal(err)
	}
	if err = f.Close(); err != nil {
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

func mustSaveCatalogFile(movies []api.Movie) {
	var err error
	var file *os.File
	if file, err = os.OpenFile(filepath.Join(testRoot, "catalog.json"), os.O_WRONLY|os.O_CREATE, 0644); err != nil {
		log.Fatal(err)
	}
	m := make(map[int]api.Movie, len(movies))
	for _, f := range movies {
		m[f.Id] = f
	}
	encoder := json.NewEncoder(file)
	if err = encoder.Encode(m); err != nil {
		log.Fatal(err)
	}
	if err = file.Close(); err != nil {
		log.Fatal(err)
	}
}

type indexMock struct {
	added []api.Movie
	found []int
}

func (idx *indexMock) Add(m api.Movie) {
	idx.added = append(idx.added, m)
}

func (idx *indexMock) Find(title string) []int {
	return idx.found
}
