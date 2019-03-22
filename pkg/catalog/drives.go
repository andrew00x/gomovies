package catalog

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrew00x/gomovies/pkg/api"
	"github.com/andrew00x/gomovies/pkg/file"
)

type drive struct {
	devSpec    string
	mountPoint string
	name       string
}

var devcd string
var etcDir string

func init() {
	devcd = "/dev"
	etcDir = "/etc"
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
	mtab := filepath.Join(etcDir, "mtab")
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

func fileDrive(drives []*drive, file string) *drive {
	return findDrive(drives, func(d *drive) bool { return strings.HasPrefix(file, d.mountPoint) })
}

func driveMounted(drives []*drive, f *api.Movie) bool {
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
