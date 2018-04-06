package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestConfigDirPathNonEmpty(t *testing.T) {
	configDir := ConfDir()
	if configDir == "" {
		t.Fatal("Returned config path directory is empty")
	}
}

func TestConfigDirIsAbsolute(t *testing.T) {
	configDir := ConfDir()
	if !filepath.IsAbs(configDir) {
		t.Fatalf("Returned path is not absolute: %s", configDir)
	}
}

func TestLoadConfig(t *testing.T) {
	dir := os.Getenv("TMPDIR")
	configPath := filepath.Join(dir, "config.json")
	createConfigFile(configPath)

	config, err := loadConfig(configPath)
	if err != nil {
		t.Fatalf("Error read config: %v", err)
	}

	expectedDirs := []string{"/home/andrew/movies"}
	expectedVideoFileExts := []string{"mkv", "avi"}
	if !reflect.DeepEqual(config.Dirs, expectedDirs) {
		t.Fatalf("Expected dirs %+v, got %+v", expectedDirs, config.Dirs)
	}
	if !reflect.DeepEqual(config.VideoFileExts, expectedVideoFileExts) {
		t.Fatalf("Expected video file extensions %+v, got %+v", expectedVideoFileExts, config.VideoFileExts)
	}
}

func createConfigFile(configPath string) {
	json := []byte("{\n" +
		"\"dirs\": [\"/home/andrew/movies\"],\n" +
		"\"video_file_exts\": [\"mkv\", \"avi\"]\n" +
		"}\n")
	ioutil.WriteFile(configPath, json, 0644)
}
