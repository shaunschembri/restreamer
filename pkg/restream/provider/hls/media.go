package hls

import (
	"context"
	"fmt"
	"time"

	"github.com/grafov/m3u8"

	"github.com/shaunschembri/restreamer/pkg/restream/provider"
	"github.com/shaunschembri/restreamer/pkg/restream/request"
)

const mbDivider = 1048576

type Media struct {
	request      request.Request
	playlistURL  string
	lastMediaSeq uint64
}

func NewMedia(request request.Request) *Media {
	return &Media{
		request: request,
	}
}

func (m Media) WithPlaylistURL(playlistURL string) *Media {
	m.playlistURL = playlistURL
	return &m
}

func (m Media) Info() string {
	return "Media"
}

func (m *Media) Get(ctx context.Context, bandwidth uint32) ([]provider.Segment, time.Duration, error) {
	playlist, err := GetPlaylist(ctx, m.request, m.playlistURL)
	if err != nil {
		return nil, 0, err
	}

	mediaPlaylist, ok := playlist.playlist.(*m3u8.MediaPlaylist)
	if !ok {
		return nil, 0, fmt.Errorf("cannot assert to a media playlist from url %s", m.playlistURL)
	}

	newSegmentsFound := false
	mediaSeq := mediaPlaylist.SeqNo
	segments := make([]provider.Segment, 0)
	for _, mediaSegment := range mediaPlaylist.Segments {
		if mediaSegment != nil {
			if mediaSeq > m.lastMediaSeq {
				newSegmentsFound = true
				m.lastMediaSeq = mediaSeq

				url, err := m.request.ResolveReference(mediaSegment.URI, playlist.referenceURL)
				if err != nil {
					return nil, 0, fmt.Errorf("cannot resolve reference URL: %w", err)
				}

				segment := provider.Segment{
					URL:       url.String(),
					KeyMethod: "NONE",
					Duration:  mediaSegment.Duration,
				}
				if mediaSegment.Key != nil {
					segment.KeyMethod = mediaSegment.Key.Method
					segment.KeyURL = mediaSegment.Key.URI
					segment.IV = mediaSegment.Key.IV
				}

				segments = append(segments, segment)
			}

			mediaSeq++
		}
	}

	// Reload playlist according to https://tools.ietf.org/html/draft-pantos-http-live-streaming-19#section-6.3.4
	reloadPlaylistAfter := time.Duration(mediaPlaylist.TargetDuration * float64(time.Second))
	if !newSegmentsFound {
		reloadPlaylistAfter /= 2
	}

	return segments, reloadPlaylistAfter, nil
}
