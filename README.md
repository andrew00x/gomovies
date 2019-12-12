# gomovies
Web interface for [omxplayer](https://github.com/popcornmix/omxplayer)

## Installation on local PC (Raspberry PI)
It requires setup Golang environment, usage of (Remote installation](#Remote installation on Raspberry PI) is preferable
* Install [omxplayer](https://github.com/popcornmix/omxplayer)
* Clone this repository to Go workspace on Raspberry
* Build
  ```
  pi@raspberrypi:~$ make
  ``` 
* Configure
  * Create in directory *$HOME/.gomovies/* file *config.json* with following content (minimal required):
    ```
    {
    "dirs": [
        "/media/pi/disk1/movies/",
        "/media/pi/disk2/movies/"
        ...
    ]
    ```
  * Supported configuration options
    * Required
      * **dirs** - list of directories with video files
    * Optional
      * **web_port** - http port, default *8000*
      * **video_file_exts** - extensions of video files, default is ```[".avi", ".mkv"]```
      * **tmdb_api_key** - api key of [The Movie Data Base (TMDb)](https://www.themoviedb.org/documentation/api). It is used for getting details about movies.
      * **tmdb_poster_small** - size of small poster, default is ```w92```, see [TMDb Images](https://developers.themoviedb.org/3/getting-started/images)
      * **tmdb_poster_large** - size of small poster, default is ```w500```, see [TMDb Images](https://developers.themoviedb.org/3/getting-started/images)
      * **torrent_remote_ctrl_addr** - address for remote control of torrent client, rtorrent is supported for now      
* Start 
  ```
  pi@raspberrypi:~$ ./gomovies
  ```
  Or in background
  ```
  pi@raspberrypi:~$ nohup ./gomovies &
  ```
* Help 
  ```
  pi@raspberrypi:~$ make help
  ```

## Remote installation on Raspberry PI
It requires [Ansible](https://www.ansible.com/)
* Install Ansible, [how](https://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html?extIdCarryOver=true&sc_cid=701f2000001OH7YAAW)
* Clone this repository
* Update file *init/ansible/raspberry.ini* with hostname of remote Raspberry PI, [how](https://www.ansible.com/overview/how-ansible-works)
* Build for Raspberry architecture. Golang installation is not needed, build will be done inside docker container
  ```
  andrew:~$ make build-rpi3
  ```
* Update configuration template, updating of **"dirs"** is needed, but can be done after installation to Raspberry.
* Install all components at once. Next will be installed: *omxplayer*, *gomovies*, *config.json*, *systemd service*. **NOTE**: Old configuration will be lost
  ```
  andrew:~$ make install-rpi3-all TMDB_API_KEY=<TMDb API access key>
  ```
  * Options:
    * TMDB_API_KEY (optional) - API key to access The Movie DB, e.g. `TMDB_API_KEY=a1b2c3d4e5f6i7`
    ```
    andrew:~$ make install-rpi3-all TMDB_API_KEY=a1b2c3d4e5f6i7
    ```
* Also all component can be installed separately.
  ```
  andrew:~$ make install-rpi3-player     # install omxplayer
  andrew:~$ make install-rpi3-bin        # install gomovies binaries
  andrew:~$ make install-rpi3-config     # install gomovies config. Old configuration will be lost. Supported option TMDB_API_KEY
  andrew:~$ make install-rpi3-systemd    # install systemd service to manage gomovies app
  ```

## Control gomovies service
* Start
 ```
 andrew:~$ make service-start
 ```
* Restart
 ```
 andrew:~$ make service-restart
 ```
* Stop
 ```
 andrew:~$ make service-stop
 ```
 
## Other
Android Application [GoMoviesDroid](https://github.com/andrew00x/GoMoviesDroid)