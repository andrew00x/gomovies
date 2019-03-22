package config

import (
	"io/ioutil"
	"log"
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

func TestConfigHasDefaultWebPort(t *testing.T) {
	dir := os.Getenv("TMPDIR")
	configPath := filepath.Join(dir, "config.json")
	mustCreateConfigFileWithContent("{}", configPath)

	config, err := loadConfig(configPath)

	assert.Nil(t, err)
	assert.Equal(t, 8000, config.WebPort)
}

func TestConfigHasDefaultVideoFileExtensions(t *testing.T) {
	dir := os.Getenv("TMPDIR")
	configPath := filepath.Join(dir, "config.json")
	mustCreateConfigFileWithContent("{}", configPath)

	config, err := loadConfig(configPath)

	assert.Nil(t, err)
	assert.Contains(t, config.VideoFileExts, ".mkv")
	assert.Contains(t, config.VideoFileExts, ".avi")
}

func TestConfigHasDefaultTMDbPosterSmall(t *testing.T) {
	dir := os.Getenv("TMPDIR")
	configPath := filepath.Join(dir, "config.json")
	mustCreateConfigFileWithContent("{}", configPath)

	config, err := loadConfig(configPath)

	assert.Nil(t, err)
	assert.Equal(t, "w92", config.TMDbPosterSmall)
}

func TestConfigHasDefaultTMDbPosterLarge(t *testing.T) {
	dir := os.Getenv("TMPDIR")
	configPath := filepath.Join(dir, "config.json")
	mustCreateConfigFileWithContent("{}", configPath)

	config, err := loadConfig(configPath)

	assert.Nil(t, err)
	assert.Equal(t, "w500", config.TMDbPosterLarge)
}

func TestConfigureConfigDirWithEnvVariable(t *testing.T) {
	dir := filepath.Join(os.Getenv("TMPDIR"), "somewhere")
	err := os.Setenv("GO_MOVIES_HOME", dir)
	assert.Nil(t, err)
	configDir := ConfDir()
	assert.Equal(t, dir, configDir)
}

func TestLoadConfig(t *testing.T) {
	json := `{
		"dirs": ["/home/andrew/movies"],
		"video_file_exts": [".mkv", ".avi", ".mp4"],
		"web_dir": "/home/andrew/gomovies/web",
		"web_port": 9999,
        "tmdb_api_key": "xyz",
        "tmdb_poster_small": "small",
        "tmdb_poster_large": "large",
        "torrent_remote_ctrl_addr": "/tmp/ctl.socket"
	}`
	dir := os.Getenv("TMPDIR")
	configPath := filepath.Join(dir, "config.json")
	mustCreateConfigFileWithContent(json, configPath)

	config, err := loadConfig(configPath)

	assert.Nil(t, err)
	assert.Equal(t, []string{"/home/andrew/movies"}, config.Dirs)
	assert.Equal(t, []string{".mkv", ".avi", ".mp4"}, config.VideoFileExts)
	assert.Equal(t, "/home/andrew/gomovies/web", config.WebDir)
	assert.Equal(t, 9999, config.WebPort)
	assert.Equal(t, "xyz", config.TMDbApiKey)
	assert.Equal(t, "small", config.TMDbPosterSmall)
	assert.Equal(t, "large", config.TMDbPosterLarge)
	assert.Equal(t, "/tmp/ctl.socket", config.TorrentRemoteCtrlAddr)
}

func mustCreateConfigFileWithContent(content, configPath string) {
	if err := ioutil.WriteFile(configPath, []byte(content), 0644); err != nil {
		log.Fatal(err)
	}
}
