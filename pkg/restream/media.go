package restream

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/grafov/m3u8"
)

type hlsPlaylist struct {
	playlist   m3u8.Playlist
	listType   m3u8.ListType
	requestURL *url.URL
}

type hlsMedia struct {
	request          request
	mediaPlaylistURL string
	lastMediaSeq     uint64
}

func (h hlsMedia) getPlaylist(ctx context.Context) (*hlsPlaylist, error) {
	response, err := h.request.do(ctx, h.mediaPlaylistURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	playlist, listType, err := m3u8.DecodeFrom(response.Body, true)
	if err != nil {
		return nil, fmt.Errorf("failed to decode playlist: %w", err)
	}

	return &hlsPlaylist{
		playlist:   playlist,
		listType:   listType,
		requestURL: response.Request.URL,
	}, nil
}

func (h hlsMedia) Info() string {
	return "Media"
}

func (h *hlsMedia) Get(ctx context.Context, bandwidth uint32) ([]segment, time.Duration, error) {
	playlist, err := h.getPlaylist(ctx)
	if err != nil {
		return nil, 0, err
	}

	mediaPlaylist, ok := playlist.playlist.(*m3u8.MediaPlaylist)
	if !ok {
		return nil, 0, fmt.Errorf("cannot assert to a media playlist from url %s", h.mediaPlaylistURL)
	}

	newSegmentsFound := false
	mediaSeq := mediaPlaylist.SeqNo
	segments := make([]segment, 0)
	for _, mediaSegment := range mediaPlaylist.Segments {
		if mediaSegment != nil {
			if mediaSeq > h.lastMediaSeq {
				newSegmentsFound = true
				h.lastMediaSeq = mediaSeq

				url, err := h.request.resolveReference(mediaSegment.URI, playlist.requestURL)
				if err != nil {
					return nil, 0, fmt.Errorf("cannot resolve reference URL: %w", err)
				}

				segment := segment{
					url:       url.String(),
					keyMethod: "NONE",
					duration:  mediaSegment.Duration,
				}
				if mediaSegment.Key != nil {
					segment.keyMethod = mediaSegment.Key.Method
					segment.keyURL = mediaSegment.Key.URI
					segment.iv = mediaSegment.Key.IV
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
