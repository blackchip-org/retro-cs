# retro-cs

[![Build Status](https://travis-ci.com/blackchip-org/retro-cs.svg?branch=master)](https://travis-ci.com/blackchip-org/retro-cs) [![GoDoc](https://godoc.org/github.com/blackchip-org/retro-cs?status.svg)](https://godoc.org/github.com/blackchip-org/retro-cs)

The Retro-Computing Systems.

Inspired by the Vintage Computer Club.

Work on the [pac8](https://github.com/blackchip-org/pac8) and the [mach85](https://github.com/blackchip-org/mach85) has now been combined into this project. Maybe [vt128](https://github.com/blackchip-org/vt128) will make it in soon. Always a work in progress and a labor of love. There is no master plan. Feel free to contact me for more information.

Click on one of the systems below for more information:

- [Commodore 64](doc/c64.md)
- [Pac-Man](doc/pacman.md)
  - and Ms. Pac-Man
- [Galaga](doc/galaga.md)

Development notes:

- [MOS Technology 6502 series processor](doc/m6502.md)
- [Pac-Man](https://github.com/blackchip-org/retro-cs/blob/master/doc/pacman.md#development-notes)
- [Zilog Z80 processor](doc/z80.md)

## Requirements

Go and SDL2 are needed to build the application.

### Linux

Install SDL with:

```bash
sudo apt-get install libsdl2{,-image,-mixer,-ttf,-gfx}-dev
```

Install go from here:

https://golang.org/dl


### macOS

Install go and SDL with:

```bash
brew install go sdl2{,_image,_mixer,_ttf,_gfx} pkg-config
```

### ROMs

ROMs are not included in this repository. Follow the directions for each system to obtain the proper ROMs or ask for the resource pack.


## Installation

```
go get github.com/blackchip-org/retro-cs/...
```

## Run

```
~/go/bin/retro-cs -s <system>
```

where `<system>` is one of the following:

- `c64`
- `galaga`
- `mspacman`
- `pacman`

Use the `-m` flag to enable the [monitor](doc/monitor.md).

## License

MIT
