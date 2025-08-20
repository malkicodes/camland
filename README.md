# CamLand

A [PixelLand](https://pixel.land) map downloader written in [Go](https://go.dev) and the successor to [PixelCam](https://github.com/malkicodes/PixelCam/)

## Installation

1. Install [Go](https://go.dev) if you haven't already.
2. Clone this repository.
3. Run `go build` and put the resulting `camland` binary wherever you store your binaries.

## Usage

To download the overworld, just run the program!

```bash
camland
```

You can also specify an output filename using the `-o` flag.

```bash
camland -o=output.png
```

To download the nether, set the `-d` flag to `nether`.

```bash
camland -d=nether
```

CamLand also has some support for custom canvases. It can only download the first layer of the first frame of a canvas, but this support will be added in the future.

```bash
camland -d=xqu69hgq
```
