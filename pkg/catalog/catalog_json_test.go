package catalog

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"testing"

	"github.com/andrew00x/gomovies/pkg/api"
	"github.com/andrew00x/gomovies/pkg/config"
)

var conf config.Config

type devDrive struct {
	name  string
	id    string
	label string
	Dev   string
	Mount string
}

var testRoot string

const sda = "usb-WDC_WD64_00AAKS-00A7B0_00A1234567E7-0:0"
const sda1Label = "wd640"
const sdb = "usb-WDC_WD50_00AAKS-00A7B0_007E7123456A-0:0"
const sdb1Label = "wd500"

var drives []devDrive
var moviesDir string
var cartoonsDir string
var movies []api.Movie

func TestLoadCatalog(t *testing.T) {
	setup()

	index := indexMock{[]indexItem{}, []int{}}
	indexFactory = func(_ *config.Config) (Index, error) { return &index, nil }

	catalog, err := createJsonCatalog(&conf)
	assert.Nil(t, err)

	expected := []api.Movie{
		{File: filepath.Join(moviesDir, "star wars", "star wars 1.avi"), Title: "star wars 1.avi", Available: true, DriveName: sda1Label},
		{File: filepath.Join(moviesDir, "star wars", "star wars 2.mkv"), Title: "star wars 2.mkv", Available: true, DriveName: sda1Label},
		{File: filepath.Join(moviesDir, "gladiator.mkv"), Title: "gladiator.mkv", Available: true, DriveName: sda1Label},
		{File: filepath.Join(moviesDir, "green mile.mkv"), Title: "green mile.mkv", Available: true, DriveName: sda1Label},
	}
	catalogContent := catalog.All()
	for i := range catalogContent { // ignore id
		catalogContent[i].Id = 0
	}
	assert.ElementsMatch(t, expected, catalogContent)

	expectedIndex := make([]indexItem, 0, len(expected))
	for _, i := range catalog.All() {
		expectedIndex = append(expectedIndex, indexItem{i.Title, i.Id})
	}
	assert.ElementsMatch(t, expectedIndex, index.added)
}

func TestCreateAndUpdateCatalog(t *testing.T) {
	setup()

	iceAge := filepath.Join(cartoonsDir, "ice age.avi")
	mustCreateFile(iceAge)
	mustSaveCatalogFile([]api.Movie{
		{File: iceAge, Title: filepath.Base(iceAge), DriveName: sdb1Label},
	})

	index := indexMock{[]indexItem{}, []int{}}
	indexFactory = func(_ *config.Config) (Index, error) { return &index, nil }

	catalog, err := createJsonCatalog(&conf)
	assert.Nil(t, err)

	expected := []api.Movie{
		{File: iceAge, Title: filepath.Base(iceAge), Available: true, DriveName: sdb1Label},
		{File: filepath.Join(moviesDir, "star wars", "star wars 1.avi"), Title: "star wars 1.avi", Available: true, DriveName: sda1Label},
		{File: filepath.Join(moviesDir, "star wars", "star wars 2.mkv"), Title: "star wars 2.mkv", Available: true, DriveName: sda1Label},
		{File: filepath.Join(moviesDir, "gladiator.mkv"), Title: "gladiator.mkv", Available: true, DriveName: sda1Label},
		{File: filepath.Join(moviesDir, "green mile.mkv"), Title: "green mile.mkv", Available: true, DriveName: sda1Label},
	}
	catalogContent := catalog.All()
	for i := range catalogContent { // ignore id
		catalogContent[i].Id = 0
	}
	assert.ElementsMatch(t, expected, catalogContent)

	expectedIndex := make([]indexItem, 0, len(expected))
	for _, i := range catalog.All() {
		expectedIndex = append(expectedIndex, indexItem{i.Title, i.Id})
	}
	assert.ElementsMatch(t, expectedIndex, index.added)
}

func TestSaveCatalog(t *testing.T) {
	setup()

	movies := map[int]*api.Movie{
		1: {File: filepath.Join(moviesDir, "star wars", "star wars 1.avi"), Title: "star wars 1.avi", Available: true},
		2: {File: filepath.Join(moviesDir, "star wars", "star wars 2.mkv"), Title: "star wars 2.mkv", Available: true},
		3: {File: filepath.Join(moviesDir, "gladiator.mkv"), Title: "gladiator.mkv", Available: true},
		4: {File: filepath.Join(moviesDir, "green mile.mkv"), Title: "green mile.mkv", Available: true},
	}
	catalog := &JsonCatalog{movies: movies}
	err := catalog.Save()
	assert.Nil(t, err)

	f, err := os.Open(catalogFile)
	assert.Nil(t, err)

	parser := json.NewDecoder(f)
	var saved map[int]*api.Movie
	err = parser.Decode(&saved)
	assert.Nil(t, err)
	err = f.Close()
	assert.Nil(t, err)

	assert.Equal(t, movies, saved)
}

func TestGetById(t *testing.T) {
	setup()

	movies := map[int]*api.Movie{
		1: {File: filepath.Join(moviesDir, "star wars", "star wars 1.avi"), Title: "star wars 1.avi", Available: true},
		2: {File: filepath.Join(moviesDir, "star wars", "star wars 2.mkv"), Title: "star wars 2.mkv", Available: true},
		3: {File: filepath.Join(moviesDir, "gladiator.mkv"), Title: "gladiator.mkv", Available: true},
		4: {File: filepath.Join(moviesDir, "green mile.mkv"), Title: "green mile.mkv", Available: true},
	}
	catalog := &JsonCatalog{movies: movies}

	result, ok := catalog.Get(1)

	assert.True(t, ok)
	assert.Equal(t, *movies[1], result)
}

func TestGetByIdReturnsEmptyResultWhenFileDoesNotExist(t *testing.T) {
	setup()
	movies := make(map[int]*api.Movie)
	catalog := &JsonCatalog{movies: movies}

	result, ok := catalog.Get(1)

	assert.False(t, ok)
	assert.Equal(t, api.Movie{}, result)
}

func TestFindByNameInCatalog(t *testing.T) {
	setup()

	movies := map[int]*api.Movie{
		1: {File: filepath.Join(moviesDir, "star wars", "star wars 1.avi"), Title: "star wars 1.avi", Available: true},
		2: {File: filepath.Join(moviesDir, "star wars", "star wars 2.mkv"), Title: "star wars 2.mkv", Available: true},
		3: {File: filepath.Join(moviesDir, "gladiator.mkv"), Title: "gladiator.mkv", Available: true},
		4: {File: filepath.Join(moviesDir, "green mile.mkv"), Title: "green mile.mkv", Available: true},
	}

	index := indexMock{added: []indexItem{}, found: []int{1, 3}}
	catalog := &JsonCatalog{movies: movies, index: &index}

	result := catalog.Find("whatever we have in index")

	expected := []api.Movie{*movies[1], *movies[3]}
	assert.Equal(t, expected, result)
}

func TestRemoveNonexistentFilesFromCatalog(t *testing.T) {
	setup()
	iceAge := filepath.Join(cartoonsDir, "ice age.avi")
	mustRemoveFiles(iceAge)
	mustSaveCatalogFile([]api.Movie{
		{File: iceAge, Title: filepath.Base(iceAge), DriveName: sdb1Label},
	})

	index := indexMock{[]indexItem{}, []int{}}
	indexFactory = func(_ *config.Config) (Index, error) { return &index, nil }

	catalog, err := createJsonCatalog(&conf)
	assert.Nil(t, err)

	expected := []api.Movie{
		{File: filepath.Join(moviesDir, "star wars", "star wars 1.avi"), Title: "star wars 1.avi", Available: true, DriveName: sda1Label},
		{File: filepath.Join(moviesDir, "star wars", "star wars 2.mkv"), Title: "star wars 2.mkv", Available: true, DriveName: sda1Label},
		{File: filepath.Join(moviesDir, "gladiator.mkv"), Title: "gladiator.mkv", Available: true, DriveName: sda1Label},
		{File: filepath.Join(moviesDir, "green mile.mkv"), Title: "green mile.mkv", Available: true, DriveName: sda1Label},
	}
	catalogContent := catalog.All()
	for i := range catalogContent { // ignore id
		catalogContent[i].Id = 0
	}
	assert.ElementsMatch(t, expected, catalogContent)

	expectedIndex := make([]indexItem, 0, len(expected))
	for _, i := range catalog.All() {
		expectedIndex = append(expectedIndex, indexItem{i.Title, i.Id})
	}
	assert.ElementsMatch(t, expectedIndex, index.added)
}

func TestKeepNonexistentFilesWhenCorrespondedDriveIsUnmounted(t *testing.T) {
	setup()
	otherMoviesDir := filepath.Join(testRoot, "media", "pi", "unmount", "movies")
	mustSaveCatalogFile([]api.Movie{
		{Id: 1, File: filepath.Join(otherMoviesDir, "rush hour 1.avi"), Title: "rush hour 1.avi", DriveName: "unmount"},
		{Id: 2, File: filepath.Join(otherMoviesDir, "rush hour 2.mkv"), Title: "rush hour 2.mkv", DriveName: "unmount"},
	})

	index := indexMock{[]indexItem{}, []int{}}
	indexFactory = func(_ *config.Config) (Index, error) { return &index, nil }

	conf := config.Config{VideoFileExts: []string{".mkv", ".avi"}, Dirs: []string{moviesDir, cartoonsDir, otherMoviesDir}}
	catalog, err := createJsonCatalog(&conf)
	assert.Nil(t, err)

	expected := []api.Movie{
		{File: filepath.Join(otherMoviesDir, "rush hour 1.avi"), Title: "rush hour 1.avi", DriveName: "unmount", Available: false},
		{File: filepath.Join(otherMoviesDir, "rush hour 2.mkv"), Title: "rush hour 2.mkv", DriveName: "unmount", Available: false},
		{File: filepath.Join(moviesDir, "star wars", "star wars 1.avi"), Title: "star wars 1.avi", DriveName: sda1Label, Available: true},
		{File: filepath.Join(moviesDir, "star wars", "star wars 2.mkv"), Title: "star wars 2.mkv", DriveName: sda1Label, Available: true},
		{File: filepath.Join(moviesDir, "gladiator.mkv"), Title: "gladiator.mkv", DriveName: sda1Label, Available: true},
		{File: filepath.Join(moviesDir, "green mile.mkv"), Title: "green mile.mkv", DriveName: sda1Label, Available: true},
	}
	catalogContent := catalog.All()
	for i := range catalogContent { // ignore id
		catalogContent[i].Id = 0
	}
	assert.ElementsMatch(t, expected, catalogContent)

	expectedIndex := make([]indexItem, 0, len(expected))
	for _, i := range catalog.All() {
		expectedIndex = append(expectedIndex, indexItem{i.Title, i.Id})
	}
	assert.ElementsMatch(t, expectedIndex, index.added)
}

func TestRefreshCatalog(t *testing.T) {
	setup()
	index := indexMock{[]indexItem{}, []int{}}
	indexFactory = func(_ *config.Config) (Index, error) { return &index, nil }

	conf := config.Config{VideoFileExts: []string{".mkv", ".avi"}, Dirs: []string{moviesDir}}
	catalog, err := createJsonCatalog(&conf)
	assert.Nil(t, err)

	for _, p := range []string{
		filepath.Join(moviesDir, "rush hour 1.avi"),
		filepath.Join(moviesDir, "rush hour 2.mkv")} {
		mustCreateFile(p)
	}
	mustRemoveMovieFiles([]api.Movie{movies[2]})

	index = indexMock{[]indexItem{}, []int{}}
	err = catalog.Refresh()
	assert.Nil(t, err)

	expected := []api.Movie{
		{File: filepath.Join(moviesDir, "rush hour 1.avi"), Title: "rush hour 1.avi", DriveName: sda1Label, Available: true},
		{File: filepath.Join(moviesDir, "rush hour 2.mkv"), Title: "rush hour 2.mkv", DriveName: sda1Label, Available: true},
		{File: filepath.Join(moviesDir, "star wars", "star wars 1.avi"), Title: "star wars 1.avi", DriveName: sda1Label, Available: true},
		{File: filepath.Join(moviesDir, "star wars", "star wars 2.mkv"), Title: "star wars 2.mkv", DriveName: sda1Label, Available: true},
		{File: filepath.Join(moviesDir, "green mile.mkv"), Title: "green mile.mkv", DriveName: sda1Label, Available: true},
	}

	catalogContent := catalog.All()
	for i := range catalogContent { // ignore id
		catalogContent[i].Id = 0
	}
	assert.ElementsMatch(t, expected, catalogContent)

	expectedIndex := make([]indexItem, 0, len(expected))
	for _, i := range catalog.All() {
		expectedIndex = append(expectedIndex, indexItem{i.Title, i.Id})
	}
	assert.ElementsMatch(t, expectedIndex, index.added)
}

func TestUpdateCatalog(t *testing.T) {
	setup()

	movies := map[int]*api.Movie{
		1: {File: filepath.Join(moviesDir, "star wars", "star wars 1.avi"), Title: "star wars 1.avi", Available: true},
		2: {File: filepath.Join(moviesDir, "star wars", "star wars 2.mkv"), Title: "star wars 2.mkv", Available: true},
		3: {File: filepath.Join(moviesDir, "gladiator.mkv"), Title: "gladiator.mkv", Available: true},
		4: {File: filepath.Join(moviesDir, "green mile.mkv"), Title: "green mile.mkv", Available: true},
	}
	catalog := &JsonCatalog{movies: movies}

	updated, err := catalog.Update(api.Movie{Id: 1, TMDbId: 101})
	assert.Nil(t, err)
	assert.Equal(t, 101, movies[1].TMDbId)
	assert.Equal(t, 101, updated.TMDbId)

	f, err := os.Open(catalogFile)
	assert.Nil(t, err)
	parser := json.NewDecoder(f)
	var saved map[int]*api.Movie
	err = parser.Decode(&saved)
	assert.Nil(t, err)
	err = f.Close()
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

func TestAddTag(t *testing.T) {
	movies := map[int]*api.Movie{
		1: {File: filepath.Join(moviesDir, "star wars", "star wars 1.avi"), Title: "star wars 1.avi", Available: true},
		2: {File: filepath.Join(moviesDir, "star wars", "star wars 2.mkv"), Title: "star wars 2.mkv", Available: true},
		3: {File: filepath.Join(moviesDir, "gladiator.mkv"), Title: "gladiator.mkv", Available: true},
		4: {File: filepath.Join(moviesDir, "green mile.mkv"), Title: "green mile.mkv", Available: true},
	}
	index := indexMock{[]indexItem{}, []int{}}
	catalog := &JsonCatalog{movies: movies, index: &index}
	err := catalog.AddTag("collection one", 2)
	assert.Nil(t, err)
	var tagged []int
	for i := range index.added {
		if index.added[i].tag == "collection one" {
			tagged = append(tagged, index.added[i].id)
		}
	}
	assert.Equal(t, []int{2}, tagged)
}

func TestAddTagFailsWhenTryTagNotExistedMovie(t *testing.T) {
	movies := map[int]*api.Movie{
		1: {File: filepath.Join(moviesDir, "star wars", "star wars 1.avi"), Title: "star wars 1.avi", Available: true},
		2: {File: filepath.Join(moviesDir, "star wars", "star wars 2.mkv"), Title: "star wars 2.mkv", Available: true},
	}
	index := indexMock{[]indexItem{}, []int{}}
	catalog := &JsonCatalog{movies: movies, index: &index}
	err := catalog.AddTag("collection one", 3)
	assert.NotNil(t, err)
	assert.Equal(t, "unable add tag for unknown movie, id: 3", err.Error())
}

func mustCreateEtcDir(rootDir string, drives []devDrive) (etc string) {
	var wd string
	var err error
	if wd, err = os.Getwd(); err != nil {
		log.Fatal(err)
	}
	var tpl *template.Template
	if tpl, err = template.New("mtab").ParseFiles(filepath.Join(wd, "testdata", "mtab")); err != nil {
		log.Fatal(err)
	}
	etc = filepath.Join(rootDir, "etc")
	mustCreateDir(etc)
	var mtabFile *os.File
	if mtabFile, err = os.OpenFile(filepath.Join(etc, "mtab"), os.O_RDWR|os.O_CREATE, 0644); err != nil {
		log.Fatal(err)
	}
	var mounts []devDrive
	for _, d := range drives {
		if d.Mount != "" {
			mounts = append(mounts, d)
		}
	}
	if err = tpl.Execute(mtabFile, mounts); err != nil {
		log.Fatal(err)
	}
	if err = mtabFile.Close(); err != nil {
		log.Fatal(err)
	}
	return
}

func setup() {
	tmp := os.Getenv("TMPDIR")
	testRoot = filepath.Join(tmp, "CatalogTest")

	if err := os.RemoveAll(testRoot); err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}
	drives = []devDrive{
		{name: "mmcblk0p0", id: "mmc-SL16G_0x2a1994a5"},
		{name: "mmcblk0p1", id: "mmc-SL16G_0x2a1994a5-part1", label: "RECOVERY", Dev: "/dev/mmcblk0p1"},
		{name: "mmcblk0p2", id: "mmc-SL16G_0x2a1994a5-part2", Dev: "/dev/mmcblk0p2"},
		{name: "mmcblk0p5", id: "mmc-SL16G_0x2a1994a5-part5", label: "SETTINGS", Dev: "/dev/mmcblk0p5"},
		{name: "mmcblk0p6", id: "mmc-SL16G_0x2a1994a5-part6", label: "boot", Dev: "/dev/mmcblk0p6"},
		{name: "mmcblk0p7", id: "mmc-SL16G_0x2a1994a5-part7", label: "root", Dev: "/dev/mmcblk0p7"},
		{name: "mmcblk0p8", id: "mmc-SL16G_0x2a1994a5-part8", label: "data", Dev: "/dev/mmcblk0p8"},
		{name: "sda", id: sda, Dev: fmt.Sprintf("/dev/%s", sda)},
		{name: "sda1", id: sda1Label, Mount: fmt.Sprintf("%s/media/pi/%s", testRoot, sda1Label), Dev: "/dev/sda1"},
		{name: "sdb", id: sdb, Dev: fmt.Sprintf("/dev/%s", sdb)},
		{name: "sdb1", id: sdb1Label, Mount: fmt.Sprintf("%s/media/pi/%s", testRoot, sdb1Label), Dev: "/dev/sdb1"}}
	devcd = mustCreateDevDir(testRoot, drives)
	etcDir = mustCreateEtcDir(testRoot, drives)
	moviesDir = filepath.Join(testRoot, "media", "pi", sda1Label, "movies")
	mustCreateDir(moviesDir)
	cartoonsDir = filepath.Join(testRoot, "media", "pi", sdb1Label, "cartoons")
	mustCreateDir(cartoonsDir)
	movies = []api.Movie{
		{File: filepath.Join(moviesDir, "star wars", "star wars 1.avi"), Title: "star wars 1.avi", Available: true},
		{File: filepath.Join(moviesDir, "star wars", "star wars 2.mkv"), Title: "star wars 2.mkv", Available: true},
		{File: filepath.Join(moviesDir, "gladiator.mkv"), Title: "gladiator.mkv", Available: true},
		{File: filepath.Join(moviesDir, "green mile.mkv"), Title: "green mile.mkv", Available: true},
	}
	mustCreateMovieFiles(movies)
	confDir := filepath.Join(testRoot, "gomovies", "config")
	mustCreateDir(confDir)
	catalogFile = filepath.Join(confDir, "catalog.json")
	conf = config.Config{VideoFileExts: []string{".mkv", ".avi"}, Dirs: []string{moviesDir, cartoonsDir}}
}

func mustCreateDevDir(rootDir string, drives []devDrive) (devDir string) {
	devDir = filepath.Join(rootDir, "dev")
	disk := filepath.Join(devDir, "disk")
	byId := filepath.Join(disk, "by-id")
	byLabel := filepath.Join(disk, "by-label")

	mustCreateDir(devDir)
	mustCreateDir(disk)
	mustCreateDir(byId)
	mustCreateDir(byLabel)

	var err error
	var wd string
	if wd, err = os.Getwd(); err != nil {
		log.Fatal(err)
	}
	for _, d := range drives {
		mustCreateFile(filepath.Join(devDir, d.name))
		if err = os.Chdir(byId); err != nil {
			log.Fatal(err)
		}
		if err = os.Symlink(filepath.Join("../..", d.name), filepath.Join(byId, d.id)); err != nil {
			log.Fatal(err)
		}
		if d.label != "" {
			if err = os.Chdir(byId); err != nil {
				log.Fatal(err)
			}
			if err = os.Symlink(filepath.Join("../..", d.name), filepath.Join(byLabel, d.label)); err != nil {
				log.Fatal(err)
			}
		}
	}
	if err = os.Chdir(wd); err != nil {
		log.Fatal(err)
	}
	return
}

func mustCreateMovieFiles(movies []api.Movie) {
	for _, m := range movies {
		mustCreateDir(filepath.Dir(m.File))
		mustCreateFile(m.File)
	}
}

func mustRemoveMovieFiles(movies []api.Movie) {
	var files []string
	for _, m := range movies {
		files = append(files, m.File)
	}
	mustRemoveFiles(files...)
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
	if file, err = os.OpenFile(catalogFile, os.O_WRONLY|os.O_CREATE, 0644); err != nil {
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

type indexItem struct {
	tag string
	id  int
}

type indexMock struct {
	added []indexItem
	found []int
}

func (idx *indexMock) Add(title string, id int) {
	idx.added = append(idx.added, indexItem{title, id})
}

func (idx *indexMock) Find(_ string) []int {
	return idx.found
}
