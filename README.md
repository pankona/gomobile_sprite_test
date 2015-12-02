# gomobile_sprite_test

This repository is for trial of gomobile and sprite.  
Goal of this activity is creating small game.

## How to build

* Install `glide` by `go get`

  * `$ go get github.com/Masterminds/glide`

  * glide is one of vendor package manager for golang.
Please visit https://github.com/Masterminds/glide to know glide.

* Export environment variable `GO15VENDOREXPERIMENT=1`

  * `export GO15VENDOREXPERIMENT=1` 
  
  * (or edit your .bashrc etc to export the environment variable automatically)

* Introduce vendor packages using glide

  * `$ glide up`

  * Then a directory for vendoring will be created automatically, and  
cloning vendor package will start.

* Build for PC Linux (or OSX)

  * `$ go build`

## License

MIT
