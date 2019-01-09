package catalog

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"
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
		movies = append(movies, api.Movie{Id: id, Path: f, Title: filepath.Base(f), DriveName: driveId})
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

func TestCreatedNewCatalog(t *testing.T) {
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
		movies = append(movies, api.Movie{Id: id, Path: f, Title: filepath.Base(f), DriveName: driveId})
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

func TestCreatedAndUpdateCatalog(t *testing.T) {
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
		movies = append(movies, api.Movie{Id: id, Path: f, Title: filepath.Base(f), DriveName: driveId})
		id++
	}

	mustSaveCatalogFile(movies)

	newFile := filepath.Join(moviesDir, "lethal weapon", "lethal weapon 4.mkv")
	mustCreateFile(newFile)
	files = append(files, newFile)
	movies = append(movies, api.Movie{Id: id, Path: newFile, Title: filepath.Base(newFile), DriveName: driveId})

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
		movies[id] = &api.Movie{Id: id, Title: filepath.Base(f), Path: f, DriveName: driveId}
		id++
	}
	catalog := &JsonCatalog{movies: movies}
	catalog.Save()

	f, err := os.Open(catalogFile)
	if err != nil {
		t.Fatalf("Unable open catalog file '%s': %v\n", catalogFile, err)
	}
	defer f.Close()
	parser := json.NewDecoder(f)
	var saved map[int]*api.Movie
	parser.Decode(&saved)

	assert.Equal(t, movies, saved)
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
		movies[id] = &api.Movie{Id: id, Path: f, Title: filepath.Base(f), DriveName: driveId}
		id++
	}

	index := indexMock{added: []api.Movie{}, found: []int{1, 3}}
	catalog := &JsonCatalog{movies: movies, index: &index}

	result := catalog.Find("whatever we have in index")

	expected := []api.Movie{*movies[1], *movies[3]}
	assert.Equal(t, expected, result)
}

func TestRemovesFromNonexistentFilesFromCatalog(t *testing.T) {
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
		movies = append(movies, api.Movie{Id: id, Path: f, Title: filepath.Base(f), DriveName: driveId})
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

func TestKeepsNonexistentFilesIfCorrespondedDriveIsUnmounted(t *testing.T) {
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
		movies = append(movies, api.Movie{Id: id, Path: f, Title: filepath.Base(f), DriveName: driveId})
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
		movies = append(movies, api.Movie{Id: id, Path: f, Title: filepath.Base(f), DriveName: driveId})
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
	movies = append(movies, api.Movie{Id: 3, Path: newFile, Title: filepath.Base(newFile), DriveName: driveId})
	index = indexMock{[]api.Movie{}, []int{}}

	catalog.Refresh()

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

type byId []api.Movie

func (m byId) Less(i, j int) bool { return m[i].Id < m[j].Id }
func (m byId) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m byId) Len() int           { return len(m) }

var moviesDir string
var driveId = "usb-WDC_WD64_00AAKS-00A7B0_00A1234567E7-0:0-part1"

var testRoot string

func setup() {
	tmp := os.Getenv("TMPDIR")
	testRoot = filepath.Join(tmp, "CatalogTest")
	err := os.RemoveAll(testRoot)
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	createDevFiles()
	etccd = filepath.Join(testRoot, "etc")
	writeMtabFile(etccd)
	moviesDir = filepath.Join(testRoot, "movies")
	mustCreateDir(moviesDir)
	catalogFile = filepath.Join(testRoot, "catalog.json")
	indexFactory = func(_ *config.Config) (Index, error) { return &indexMock{}, nil }
}

func writeMtabFile(etc string) {
	mustCreateDir(etc)
	mtab := filepath.Join(etc, "mtab")
	mtabContent := []byte(`/dev/root / ext4 rw,noatime,data=ordered 0 0
devtmpfs /dev devtmpfs rw,relatime,size=468148k,nr_inodes=117037,mode=755 0 0
sysfs /sys sysfs rw,nosuid,nodev,noexec,relatime 0 0
proc /proc proc rw,relatime 0 0
tmpfs /dev/shm tmpfs rw,nosuid,nodev 0 0
devpts /dev/pts devpts rw,nosuid,noexec,relatime,gid=5,mode=620,ptmxmode=000 0 0
tmpfs /run tmpfs rw,nosuid,nodev,mode=755 0 0
tmpfs /run/lock tmpfs rw,nosuid,nodev,noexec,relatime,size=5120k 0 0
tmpfs /sys/fs/cgroup tmpfs ro,nosuid,nodev,noexec,mode=755 0 0
cgroup /sys/fs/cgroup/systemd cgroup rw,nosuid,nodev,noexec,relatime,xattr,release_agent=/lib/systemd/systemd-cgroups-agent,name=systemd 0 0
cgroup /sys/fs/cgroup/memory cgroup rw,nosuid,nodev,noexec,relatime,memory 0 0
cgroup /sys/fs/cgroup/freezer cgroup rw,nosuid,nodev,noexec,relatime,freezer 0 0
cgroup /sys/fs/cgroup/net_cls cgroup rw,nosuid,nodev,noexec,relatime,net_cls 0 0
cgroup /sys/fs/cgroup/devices cgroup rw,nosuid,nodev,noexec,relatime,devices 0 0
cgroup /sys/fs/cgroup/cpu,cpuacct cgroup rw,nosuid,nodev,noexec,relatime,cpu,cpuacct 0 0
cgroup /sys/fs/cgroup/blkio cgroup rw,nosuid,nodev,noexec,relatime,blkio 0 0
systemd-1 /proc/sys/fs/binfmt_misc autofs rw,relatime,fd=28,pgrp=1,timeout=0,minproto=5,maxproto=5,direct 0 0
mqueue /dev/mqueue mqueue rw,relatime 0 0
debugfs /sys/kernel/debug debugfs rw,relatime 0 0
sunrpc /run/rpc_pipefs rpc_pipefs rw,relatime 0 0
configfs /sys/kernel/config configfs rw,relatime 0 0
/dev/mmcblk0p6 /boot vfat rw,relatime,fmask=0022,dmask=0022,codepage=437,iocharset=ascii,shortname=mixed,errors=remount-ro 0 0
tmpfs /run/user/1000 tmpfs rw,nosuid,nodev,relatime,size=94548k,mode=700,uid=1000,gid=1000 0 0
gvfsd-fuse /run/user/1000/gvfs fuse.gvfsd-fuse rw,nosuid,nodev,relatime,user_id=1000,group_id=1000 0 0
fusectl /sys/fs/fuse/connections fusectl rw,relatime 0 0
/dev/mmcblk0p8 /media/pi/data ext4 rw,nosuid,nodev,relatime,data=ordered 0 0
/dev/mmcblk0p5 /media/pi/SETTINGS ext4 rw,nosuid,nodev,relatime,data=ordered 0 0
` + "/dev/sda1 " + filepath.Join(testRoot, "movies") + " ntfs ro,nosuid,nodev,relatime,uid=1000,gid=1000,fmask=0177,dmask=077,nls=utf8,errors=continue,mft_zone_multiplier=1 0 0\n")

	ioutil.WriteFile(mtab, mtabContent, 0644)
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
		"sda":       {id: "usb-WDC_WD64_00AAKS-00A7B0_00A1234567E7-0:0"},
		"sda1":      {id: driveId}}
	wd, _ := os.Getwd()
	for n, d := range drives {
		mustCreateFile(filepath.Join(dev, n))
		os.Chdir(byId)
		os.Symlink(filepath.Join("../..", n), filepath.Join(byId, d.id))
		if d.label != "" {
			os.Chdir(byId)
			os.Symlink(filepath.Join("../..", n), filepath.Join(byLabel, d.label))
		}
	}
	os.Chdir(wd)
	devcd = dev
}

func mustCreateDir(dir string) {
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		panic(err)
	}
}

func mustCreateFile(path string) {
	f, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	err = f.Close()
	if err != nil {
		panic(err)
	}
}

func mustRemoveFiles(paths ...string) {
	for _, path := range paths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			panic(err)
		}
	}
}

func mustSaveCatalogFile(movies []api.Movie) {
	file, err := os.OpenFile(filepath.Join(testRoot, "catalog.json"), os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	m := make(map[int]api.Movie, len(movies))
	for _, f := range movies {
		m[f.Id] = f
	}
	encoder := json.NewEncoder(file)
	encoder.Encode(m)
	err = file.Close()
	if err != nil {
		panic(err)
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
