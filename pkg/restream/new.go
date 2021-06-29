package restream

import (
	"context"
	"io"

	"github.com/shaunschembri/restreamer/pkg/restream/provider"
)

const (
	defaultUserAgent      = "restreamer"
	defaultBandwidth      = 10485760
	defaultReadBufferSize = 1048576
	mbDivider             = 1048576
)

type Restream struct {
	UserAgent        string
	MaxBandwidth     uint32
	Writer           io.Writer
	SegmentProvider  provider.Provider
	ReadBufferSize   int
	streamedBytes    int64
	currentBandwidth uint32
	segments         chan provider.Segment
	errors           chan error
	decrypter        decrypter
}

func (r *Restream) init(ctx context.Context, playlistURL string) error {
	r.segments = make(chan provider.Segment, 1024)
	r.errors = make(chan error, 1024)

	if r.MaxBandwidth == 0 {
		r.MaxBandwidth = defaultBandwidth
	}
	r.currentBandwidth = r.MaxBandwidth

	if r.ReadBufferSize == 0 {
		r.ReadBufferSize = defaultReadBufferSize
	}

	if r.UserAgent == "" {
		r.UserAgent = defaultUserAgent
	}

	if r.SegmentProvider == nil {
		segmentProvider, err := r.detectStream(ctx, playlistURL, r.UserAgent, r.MaxBandwidth)
		if err != nil {
			return err
		}

		r.SegmentProvider = segmentProvider
	}

	return nil
}
