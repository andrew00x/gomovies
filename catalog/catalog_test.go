package catalog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"github.com/andrew00x/gomovies/config"
)

func TestLoadCatalog(t *testing.T) {
	cleanupCatalog()
	expected := map[int]*MovieFile{
		1000: {Id: 1000, Title: "xxx.mkv", Path: filepath.Join(testRoot, "movies", "xxx.mkv"), DriveName: "usb-WDC_WD64_00AAKS-00A7B0_00A1234567E7-0:0-part1"},
		1001: {Id: 1001, Title: "yyy.avi", Path: filepath.Join(testRoot, "movies", "yyy.avi"), DriveName: "usb-WDC_WD64_00AAKS-00A7B0_00A1234567E7-0:0-part1"},
	}
	mustSaveCatalogFile(expected)

	conf := &config.Config{CatalogFile: filepath.Join(testRoot, "catalog.json")}
	catalog, err := NewCatalog(conf)
	if err != nil {
		t.Fatalf("Unable create catalog: %v\n", err)
	}
	if !reflect.DeepEqual(expected, catalog.movies) {
		t.Fatalf("Expected: %v\nbut actual is: %v\n", expected, catalog.movies)
	}
}

func TestCreatedNewCatalog(t *testing.T) {
	cleanupCatalog()

	expected := make(map[int]*MovieFile, len(movieFiles))
	var id = 0
	for _, path := range movieFiles {
		id++
		expected[id] = &MovieFile{Id: id, Path: path, Title: filepath.Base(path), DriveName: driveId}
	}

	conf := &config.Config{CatalogFile: filepath.Join(testRoot, "catalog.json"), Dirs: []string{testRoot}, VideoFileExts: []string{".mkv", ".avi"}}
	catalog, err := NewCatalog(conf)
	if err != nil {
		t.Fatalf("Unable create catalog: %v\n", err)
	}

	if !reflect.DeepEqual(expected, catalog.movies) {
		t.Fatalf("Expected: %v\nbut actual is: %v\n", expected, catalog.movies)
	}
}

func TestCreatedAndUpdateCatalog(t *testing.T) {
	cleanupCatalog()

	expected := map[int]*MovieFile{
		1000: {Id: 1000, Title: "xxx.mkv", Path: filepath.Join(testRoot, "movies", "xxx.mkv"), DriveName: "usb-WDC_WD64_00AAKS-00A7B0_00A1234567E7-0:0-part1"},
		1001: {Id: 1001, Title: "yyy.avi", Path: filepath.Join(testRoot, "movies", "yyy.avi"), DriveName: "usb-WDC_WD64_00AAKS-00A7B0_00A1234567E7-0:0-part1"},
	}
	mustSaveCatalogFile(expected)

	var id = 1001
	for _, path := range movieFiles {
		id++
		expected[id] = &MovieFile{Id: id, Path: path, Title: filepath.Base(path), DriveName: driveId}
	}

	conf := &config.Config{CatalogFile: filepath.Join(testRoot, "catalog.json"), Dirs: []string{testRoot}, VideoFileExts: []string{".mkv", ".avi"}}
	catalog, err := NewCatalog(conf)
	if err != nil {
		t.Fatalf("Unable create catalog: %v\n", err)
	}

	if !reflect.DeepEqual(expected, catalog.movies) {
		t.Fatalf("Expected: %v\nbut actual is: %v\n", expected, catalog.movies)
	}
}

func TestSaveCatalog(t *testing.T) {
	cleanupCatalog()

	movies := map[int]*MovieFile{
		1000: {Id: 1000, Title: "xxx.mkv", Path: filepath.Join(testRoot, "movies", "xxx.mkv"), DriveName: "usb-WDC_WD64_00AAKS-00A7B0_00A1234567E7-0:0-part1"},
		1001: {Id: 1001, Title: "yyy.avi", Path: filepath.Join(testRoot, "movies", "yyy.avi"), DriveName: "usb-WDC_WD64_00AAKS-00A7B0_00A1234567E7-0:0-part1"},
	}
	catalogFile := filepath.Join(testRoot, "catalog.json")
	catalog := &Catalog{movies: movies, catalogFile: catalogFile}
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

	var id = 7001
	all := make(map[int]*MovieFile, len(movieFiles))
	for _, path := range movieFiles {
		id++
		all[id] = &MovieFile{Id: id, Path: path, Title: filepath.Base(path), DriveName: driveId}
	}
	mustSaveCatalogFile(all)

	conf := &config.Config{CatalogFile: filepath.Join(testRoot, "catalog.json"), Dirs: []string{testRoot}, VideoFileExts: []string{".mkv", ".avi"}}
	catalog, err := NewCatalog(conf)
	if err != nil {
		t.Fatalf("Unable create catalog: %v\n", err)
	}

	expected := make(map[int]MovieFile)
	for _, m := range all {
		if strings.Contains(m.Title, "mkv") {
			expected[m.Id] = *m
		}
	}

	result := make(map[int]MovieFile)
	for _, m := range catalog.Find("mkV") {
		result[m.Id] = m
	}

	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("Expected: %v\nbut actual is: %v\n", expected, result)
	}
}

func TestGetAllInCatalog(t *testing.T) {
	cleanupCatalog()

	var id = 9001
	all := make(map[int]*MovieFile, len(movieFiles))
	for _, path := range movieFiles {
		id++
		all[id] = &MovieFile{Id: id, Path: path, Title: filepath.Base(path), DriveName: driveId}
	}
	mustSaveCatalogFile(all)

	conf := &config.Config{CatalogFile: filepath.Join(testRoot, "catalog.json"), Dirs: []string{testRoot}, VideoFileExts: []string{".mkv", ".avi"}}
	catalog, err := NewCatalog(conf)
	if err != nil {
		t.Fatalf("Unable create catalog: %v\n", err)
	}

	expected := make(map[int]MovieFile)
	for _, m := range all {
		expected[m.Id] = *m
	}

	result := make(map[int]MovieFile)
	for _, m := range catalog.All() {
		result[m.Id] = m
	}

	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("Expected: %v\nbut actual is: %v\n", expected, result)
	}
}
