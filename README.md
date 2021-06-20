# Restreamer

Simple golang application that downloads the media segments from an [HTTP Live Streaming (HLS)](https://en.wikipedia.org/wiki/HTTP_Live_Streaming) and streams the segments over an HTTP connection or stores them to disk as a single file.  It is written in Go which makes it extremely portable and uses little system resources which makes it ideal to be used on embedded devices where resources are limited (ex TV set top boxes).

This application enables media players with limited or no support for HLS to play these streams as many media players are able to play a stream sourced through a single HTTP connection. Alternatively the stream can be saved to a storage device and viewed as any other media file, possibly even while downloading as a time-shifted programme.

## Features
- Support non-encrypted and AES128 encrypted streams
- Automatically detects if the M3U8 contains a master or media playlist
- Automatic selection of a stream variant from the master playlist depending on the available bandwidth

## Quick Start Guide
- Download `restreamer` binary for you target system. Pre-build binaries are available [here](https://github.com/shaunschembri/restreamer/releases) alternatively build from source following the [Building restreamer](#building-restreamer) section.
- Edit [restreamer.yaml](configs/restreamer.yaml) and add the M3U8 URLs for the streams you like to use, giving each URL a unique stream id.
- Save the config on the target system.  By default the config file is expected to be in  `$HOME\.restreamer` but this can be stored in any accessible path and passed using the `--config` argument.

### Using restreamer over HTTP
- Execute `restreamer server`
- Initiate playback on your media player of choice by streaming from `http://ip-address:port/stream-id` Example `http://localhost:1230/nasatv1`

Available options for `server` sub-command are

```
  -h, --help                  help for server
  -a, --http-address string   http server bind address (default "127.0.0.1")
  -p, --http-port int         http server listening port (default 1230)
```

### Using restreamer to download to local storage
Execute `restreamer download -s nasatv1 -t 1h` which would stream the channel for 1 hour and store all segments as a single file in the path provided.

WARNING: While there is no technical limitation to use `restreamer` over the public internet, this is strongly not recommended due to the lack of secure transport (HTTPS) and authentication in the implementation.  However, it is very possible to re-stream over a LAN where the lack of security is not an issue. 

Available options for download sub-command are

```
  -d, --download-path string   path to store downloaded media (default ".")
  -t, --duration duration      stream duration (default 12h0m0s)
  -f, --filename string        filename of downloaded media
  -h, --help                   help for download
  -s, --stream-id string       stream id
```

### Global flags
Both of the sub-commands can also control some parameters of `restreamer` library.  These commands are

```
  -c, --config string         config file
  -m, --max-bandwidth float   max bandwidth in mb/sec (default 10)
  -b, --read-buffer float     read buffer in mb (default 1)
```

## Building restreamer
- Install go `v1.16` or later. You can obtain the binaries for you operating from [here](https://golang.org/dl/)
- Clone this repo with `git clone https://github.com/shaunschembri/restreamer`
- Execute `go build -o restreamer cmd/restreamer/main.go` to create a binary that can be executed on your local machine.
- To build for a different system set the `GOOS`, `GOARCH` and other target specific environmental variables before executing the above command.  The complete list of valid combinations can be found [here](https://golang.org/doc/install/source#environment).  Example to build for an ARM linux set-top box you need to execute `GOOS=linux GOARCH=arm GOARM=6 go build...`

## Using the library in your code

All the heavy lifting happens in the [restreamer](https://github.com/shaunschembri/restreamer/tree/main/pkg/restream) package and the stream is returned to the calling code through an `io.Writer` so it can be easily consumed by other applications.  While the [restreamer app](https://github.com/shaunschembri/restreamer/tree/main/internal/restreamer) is a complete and simple enough example an ever simpler example to download a stream to disk is shown below

```go
package main

import (
	"context"
	"log"
	"os"

	"github.com/shaunschembri/restreamer/pkg/restream"
)

func main() {
	file, err := os.Create("nasatv1.ts")
	if err != nil {
		log.Fatalln("Error: cannot open file nasatv1.ts for writing")
	}

	restreamer := restream.Restream{
		Writer: file,
	}

	if err := restreamer.Start(context.Background(), "https://ntv1.akamaized.net/hls/live/2014075/NASA-NTV1-HLS/master.m3u8"); err != nil {
		log.Fatalln(err)
	}
}
```

## Future work
- Support remuxing of the output stream, making it possible to add subtitles and audio streams provided through separate segments.
- Support other [Adaptive Bitrate Streaming](https://en.wikipedia.org/wiki/Adaptive_bitrate_streaming) systems like [MPEG-DASH](https://en.wikipedia.org/wiki/Dynamic_Adaptive_Streaming_over_HTTP). The code has been on propose developed to be generic enough to support other systems that break down the video stream in multiple segments.
- Support `SAMPLE-AES` encryption, provided a good example not tied with a proprietary DRM system is available.
- Cover all code with a comprehensive test suite.

## License
Licensed under the [3-Clause BSD License](LICENSE.txt)