package catalog

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"os"
	"testing"
	"sort"
	"github.com/andrew00x/gomovies/config"
)

func TestLoadCatalog(t *testing.T) {
	cleanupCatalog()
	expected := make([]MovieFile, 0, len(movieFiles))
	var id = 0
	for _, path := range movieFiles {
		id++
		expected = append(expected, MovieFile{Id: id, Path: path, Title: filepath.Base(path), DriveName: driveId})
	}
	mustSaveCatalogFile(expected)

	conf := &config.Config{CatalogFile: filepath.Join(testRoot, "catalog.json")}
	catalog, err := Create(conf)
	if err != nil {
		t.Fatalf("Unable create catalog: %v\n", err)
	}
	result := catalog.All()
	sort.Sort(ById(result))
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("Expected: %v\nbut actual is: %v\n", expected, result)
	}
}
func TestCreatedNewCatalog(t *testing.T) {
	cleanupCatalog()

	expected := make([]MovieFile, 0, len(movieFiles))
	var id = 0
	for _, path := range movieFiles {
		id++
		expected = append(expected, MovieFile{Id: id, Path: path, Title: filepath.Base(path), DriveName: driveId})
	}

	conf := &config.Config{CatalogFile: filepath.Join(testRoot, "catalog.json"), Dirs: []string{testRoot}, VideoFileExts: []string{".mkv", ".avi"}}
	catalog, err := Create(conf)
	if err != nil {
		t.Fatalf("Unable create catalog: %v\n", err)
	}

	result := catalog.All()
	sort.Sort(ById(result))
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("Expected: %v\nbut actual is: %v\n", expected, result)
	}
}

func TestCreatedAndUpdateCatalog(t *testing.T) {
	cleanupCatalog()
	mustCreateFile(filepath.Join(testRoot, "movies"), "xxx.mkv")
	defer func() {
		os.Remove(filepath.Join(testRoot, "movies", "xxx.mkv"))
	}()

	var id = 0
	expected := []MovieFile{
		{Id: id, Title: "xxx.mkv", Path: filepath.Join(testRoot, "movies", "xxx.mkv"), DriveName: "usb-WDC_WD64_00AAKS-00A7B0_00A1234567E7-0:0-part1"},
	}
	mustSaveCatalogFile(expected)

	for _, path := range movieFiles {
		id++
		expected = append(expected, MovieFile{Id: id, Path: path, Title: filepath.Base(path), DriveName: driveId})
	}

	conf := &config.Config{CatalogFile: filepath.Join(testRoot, "catalog.json"), Dirs: []string{testRoot}, VideoFileExts: []string{".mkv", ".avi"}}
	catalog, err := Create(conf)
	if err != nil {
		t.Fatalf("Unable create catalog: %v\n", err)
	}

	result := catalog.All()
	sort.Sort(ById(result))
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("Expected: %v\nbut actual is: %v\n", expected, result)
	}
}

func TestSaveCatalog(t *testing.T) {
	cleanupCatalog()

	movies := map[int]*MovieFile{
		1: {Id: 1, Title: "xxx.mkv", Path: filepath.Join(testRoot, "movies", "xxx.mkv"), DriveName: "usb-WDC_WD64_00AAKS-00A7B0_00A1234567E7-0:0-part1"},
		2: {Id: 2, Title: "yyy.avi", Path: filepath.Join(testRoot, "movies", "yyy.avi"), DriveName: "usb-WDC_WD64_00AAKS-00A7B0_00A1234567E7-0:0-part1"},
	}
	catalogFile := filepath.Join(testRoot, "catalog.json")
	catalog := &FSCatalog{movies: movies, catalogFile: catalogFile}
	catalog.Save()

	f, err := os.Open(catalogFile)
	if err != nil {
		t.Fatalf("Unable open catalog file '%s': %v\n", catalogFile, err)
	}
	defer f.Close()
	parser := json.NewDecoder(f)
	var saved map[int]*MovieFile
	parser.Decode(&saved)

	if !reflect.DeepEqual(movies, saved) {
		t.Fatalf("Expected: %v\nbut actual is: %v\n", movies, saved)
	}
}

func TestFindByNameInCatalog(t *testing.T) {
	cleanupCatalog()

	var id = 0
	all := make([]MovieFile, 0, len(movieFiles))
	for _, path := range movieFiles {
		id++
		all = append(all, MovieFile{Id: id, Path: path, Title: filepath.Base(path), DriveName: driveId})
	}
	mustSaveCatalogFile(all)

	conf := &config.Config{CatalogFile: filepath.Join(testRoot, "catalog.json"), Dirs: []string{testRoot}, VideoFileExts: []string{".mkv", ".avi"}}
	catalog, err := Create(conf)
	if err != nil {
		t.Fatalf("Unable create catalog: %v\n", err)
	}

	expected := make([]MovieFile, 0, 2)
	for _, m := range all {
		if strings.Contains(m.Title, "mkv") {
			expected = append(expected, m)
		}
	}

	result := catalog.Find("mkV")
	sort.Sort(ById(result))
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("Expected: %v\nbut actual is: %v\n", expected, result)
	}
}

func TestRemovesFromNonexistentFilesFromCatalog(t *testing.T) {
	cleanupCatalog()

	var id = 0
	all := make([]MovieFile, 0, len(movieFiles)+1)
	for _, path := range movieFiles {
		id++
		all = append(all, MovieFile{Id: id, Path: path, Title: filepath.Base(path), DriveName: driveId})
	}
	nonexistent := filepath.Join(testRoot, "movies", "x", "nonexistent.mkv")
	id++
	all = append(all, MovieFile{Id: id, Path: nonexistent, Title: filepath.Base(nonexistent), DriveName: driveId})

	mustSaveCatalogFile(all)

	conf := &config.Config{CatalogFile: filepath.Join(testRoot, "catalog.json"), Dirs: []string{testRoot}, VideoFileExts: []string{".mkv", ".avi"}}
	catalog, err := Create(conf)
	if err != nil {
		t.Fatalf("Unable create catalog: %v\n", err)
	}

	for _, m := range catalog.All() {
		if m.Path == nonexistent {
			t.Fatalf("File %s must be removed from catalog since it does not exist", m.Path)
		}
	}
}

func TestKeepsNonexistentFilesIfCorrespondedDriveIsUnmounted(t *testing.T) {
	cleanupCatalog()

	var id = 0
	all := make([]MovieFile, 0, len(movieFiles))
	for _, path := range movieFiles {
		id++
		all = append(all, MovieFile{Id: id, Path: path, Title: filepath.Base(path), DriveName: driveId})
	}
	nonexistent := filepath.Join(testRoot, "movies2", "x", "nonexistent.mkv")
	id++
	all = append(all, MovieFile{Id: id, Path: nonexistent, Title: filepath.Base(nonexistent), DriveName: "unmounted-drive"})

	mustSaveCatalogFile(all)

	conf := &config.Config{CatalogFile: filepath.Join(testRoot, "catalog.json"), Dirs: []string{testRoot}, VideoFileExts: []string{".mkv", ".avi"}}
	catalog, err := Create(conf)
	if err != nil {
		t.Fatalf("Unable create catalog: %v\n", err)
	}

	result := catalog.All()
	sort.Sort(ById(result))
	if !reflect.DeepEqual(all, result) {
		t.Fatalf("Expected: %v\nbut actual is: %v\n", all, result)
	}
}

func TestRefreshCatalog(t *testing.T) {
	cleanupCatalog()

	conf := &config.Config{CatalogFile: filepath.Join(testRoot, "catalog.json"), Dirs: []string{}, VideoFileExts: []string{".mkv", ".avi"}}
	catalog, err := Create(conf)
	if err != nil {
		t.Fatalf("Unable create catalog: %v\n", err)
	}

	if len(catalog.All()) != 0 {
		t.Fatal("catalog expected to be empty")
	}

	conf.Dirs = []string{testRoot}
	catalog.Refresh(conf)

	afterRefresh := catalog.All()
	m := make(map[string]bool)
	for _, f := range afterRefresh {
		m[f.Path] = true
	}
	for _, f := range movieFiles {
		if !m[f] {
			t.Fatalf("file %s expected to be in catalog after refresh", f)
		}
	}
}
