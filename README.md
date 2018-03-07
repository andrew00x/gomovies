# gomovies
Web interface for [omxplayer](https://github.com/popcornmix/omxplayer)

## Pre-requirements
Install omxd [omxplayer](https://github.com/popcornmix/omxplayer)

## Configurations
Create in directory *$HOME/.gomovies/* file *config.json* with following content:
```
{
"dirs": [
    "/media/pi/disk1/movies/",
    "/media/pi/disk2/movies/"
    ...
],
"video_file_exts": [
    ".avi",
    ".mkv",
    ...
]
```
### Supported configuration options
* **dirs** - list of directories with video files
* **video_file_exts** - extensions of video files

#### Optional
* **catalog_file** - path to json file where gomovies stores list of all known video files, default is *$HOME/.gomovies/catalog.json*
* **web_port** - http port, default *8000*
* **web_dir** - path to directory that contains web client, for instance [gomovies-react](https://github.com/andrew00x/gomovies-react)

## Installation
Clone repository to your Go workspace and run:
```
pi@raspberrypi:~$ go install github.com/andrew00x/gomovies
``` 
And start it:
```gomovies```