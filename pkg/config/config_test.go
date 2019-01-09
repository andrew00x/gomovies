package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
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

func TestConfigHasDefaultVideoFileExtensions(t *testing.T) {
	dir := os.Getenv("TMPDIR")
	configPath := filepath.Join(dir, "config.json")
	createConfigFileWithContent("{}", configPath)

	config, err := loadConfig(configPath)

	assert.Nil(t, err)
	assert.Contains(t, config.VideoFileExts, ".mkv")
	assert.Contains(t, config.VideoFileExts, ".avi")
}

func TestConfigureConfigDirWithEnvVariable(t *testing.T) {
	dir := filepath.Join(os.Getenv("TMPDIR"), "somewhere")
	os.Setenv("GO_MOVIES_HOME", dir)
	configDir := ConfDir()
	assert.Equal(t, dir, configDir)
}

func TestLoadConfig(t *testing.T) {
	json := "{\n" +
		"\"dirs\": [\"/home/andrew/movies\"],\n" +
		"\"video_file_exts\": [\".mkv\", \".avi\", \".mp4\"]\n" +
		"}\n"
	dir := os.Getenv("TMPDIR")
	configPath := filepath.Join(dir, "config.json")
	createConfigFileWithContent(json, configPath)

	config, err := loadConfig(configPath)

	assert.Nil(t, err)
	assert.Equal(t, []string{"/home/andrew/movies"}, config.Dirs)
	assert.Contains(t, config.VideoFileExts, ".mkv")
	assert.Contains(t, config.VideoFileExts, ".avi")
	assert.Contains(t, config.VideoFileExts, ".mp4")
}

func createConfigFileWithContent(content, configPath string) {
	ioutil.WriteFile(configPath, []byte(content), 0644)
}
