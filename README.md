# gomovies
Web interface for [omxplayer](https://github.com/popcornmix/omxplayer)

## Pre-requirements
Install [omxplayer](https://github.com/popcornmix/omxplayer)

## Configurations
Create in directory *$HOME/.gomovies/* file *config.json* with following content:
```
{
"dirs": [
    "/media/pi/disk1/movies/",
    "/media/pi/disk2/movies/"
    ...
]
```
### Supported configuration options

#### Required
* **dirs** - list of directories with video files [Installation]

#### Optional
* **catalog_file** - path to json file where gomovies stores list of all known video files, default is *$HOME/.gomovies/catalog.json*
* **web_port** - http port, default *8000*
* **web_dir** - path to directory that contains web client, for instance [gomovies-react](https://github.com/andrew00x/gomovies-react)
* **video_file_exts** - extensions of video files, default is ```[".avi", ".mkv"]```
* **tmdb_api_key** - api key of [The Movie Data Base (TMDb)](https://www.themoviedb.org/documentation/api). It is used for getting details about movies.
* **tmdb_poster_small** - size of small poster, default is ```w92```, see [TMDb Images](https://developers.themoviedb.org/3/getting-started/images)
* **tmdb_poster_large** - size of small poster, default is ```w500```, see [TMDb Images](https://developers.themoviedb.org/3/getting-started/images)

## Installation
Clone repository to your Go workspace and run:
```
pi@raspberrypi:~$ make
``` 
And start it:
```./gomovies``` or in background ```nohup ./gomovies &```

Run `make help` for details.
