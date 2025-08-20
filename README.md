# CamLand

A [PixelLand](https://pixel.land) map downloader written in [Go](https://go.dev) and the successor to [PixelCam](https://github.com/malkicodes/PixelCam/)

## Installation

1. Install [Go](https://go.dev) if you haven't already.
2. Get the binary by either:
    - Cloning the repository and running `go build`, or
    - Downloading the latest binary [here](https://github.com/malkicodes/camland/releases/latest)

## Usage

To download the overworld, just run the program!

```bash
camland
```

You can also specify an output filename or directory using the `-o` flag.

```bash
camland -o=output.png # Saves to output.png
camland -o=images/ # Saves to images/camland_overworld_<TIMESTAMP>.png
```

To download the nether, set the `-d` flag to `nether`.

```bash
camland -d=nether
```

CamLand also has some support for custom canvases.

```bash
camland -d=xqu69hgq
camland -d=s8lpvkwy -gif # Downloads as a GIF
```
