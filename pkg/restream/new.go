package restream

import (
	"context"
	"io"
	"time"
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
	SegmentProvider  SegmentProvider
	ReadBufferSize   int
	streamedBytes    int64
	currentBandwidth uint32
	segments         chan segment
	errors           chan error
	decrypter        decrypter
}

type segment struct {
	url       string
	keyMethod string
	keyURL    string
	iv        string
	duration  float64
}

type SegmentProvider interface {
	Get(ctx context.Context, bandwidth uint32) ([]segment, time.Duration, error)
	Info() string
}

func (r *Restream) init(ctx context.Context, playlistURL string) error {
	r.segments = make(chan segment, 1024)
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
		if err := r.getSegmentProvider(ctx, playlistURL); err != nil {
			return err
		}
	}

	return nil
}
