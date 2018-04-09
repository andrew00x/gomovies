package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestConfigDirPathNonEmpty(t *testing.T) {
	configDir := ConfDir()
	assert.NotEmpty(t, configDir)
}

func TestConfigDirIsAbsolute(t *testing.T) {
	configDir := ConfDir()
	assert.True(t, filepath.IsAbs(configDir))
}

func TestConfigureConfigDirWithEnvVariable(t *testing.T) {
	dir := filepath.Join(os.Getenv("TMPDIR"), "somewhere")
	os.Setenv("GO_MOVIES_HOME", dir)
	configDir := ConfDir()
	assert.Equal(t, dir, configDir)
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
