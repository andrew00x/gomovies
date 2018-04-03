package catalog

import (
	"testing"
	"os"
	"path/filepath"
	"io/ioutil"
	"encoding/json"
)

type ById []MovieFile

func (m ById) Less(i, j int) bool { return m[i].Id < m[j].Id }
func (m ById) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m ById) Len() int           { return len(m) }


func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

var driveId = "usb-WDC_WD64_00AAKS-00A7B0_00A1234567E7-0:0-part1"
var testRoot string
var movieFiles []string

func setup() {
	tmp := os.Getenv("TMPDIR")
	testRoot = filepath.Join(tmp, "IndexTest")
	err := os.RemoveAll(testRoot)
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}

	createDevFiles()
	createMovieFiles()

	etccd = filepath.Join(testRoot, "etc")
	writeMtabFile(etccd)
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

func createMovieFiles() {
	movies := filepath.Join(testRoot, "movies")

	mustCreateDir(filepath.Join(movies, "a", "b", "c"))
	mustCreateDir(filepath.Join(movies, "j", "k", "l"))

	mustCreateFile(filepath.Join(movies, "a"), "a.avi")
	mustCreateFile(filepath.Join(movies, "a", "b", "c"), "c.mkv")
	mustCreateFile(filepath.Join(movies, "j"), "j.mkv")
	mustCreateFile(filepath.Join(movies, "j", "k", "l"), "l.avi")

	movieFiles = []string{filepath.Join(movies, "a", "a.avi"),
		filepath.Join(movies, "a", "b", "c", "c.mkv"),
		filepath.Join(movies, "j", "j.mkv"),
		filepath.Join(movies, "j", "k", "l", "l.avi")}
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
		mustCreateFile(dev, n)
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

func mustCreateFile(dir string, name string) {
	f, err := os.OpenFile(filepath.Join(dir, name), os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	err = f.Close()
	if err != nil {
		panic(err)
	}
}

func mustSaveCatalogFile(movies []MovieFile) {
	file, err := os.OpenFile(filepath.Join(testRoot, "catalog.json"), os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	encoder := json.NewEncoder(file)
	encoder.Encode(toMapWithIdAsKey(movies))
	err = file.Close()
	if err != nil {
		panic(err)
	}
}

func toMapWithIdAsKey(movies []MovieFile) map[int]MovieFile{
	m := make(map[int]MovieFile, len(movies))
	for _, f := range movies {
		m[f.Id] = f
	}
	return m
}



func cleanupCatalog() {
	catalogFile := filepath.Join(testRoot, "catalog.json")
	if err := os.Remove(catalogFile); err != nil && !os.IsNotExist(err) {
		panic(err)
	}
}
