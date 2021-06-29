package hls

import (
	"context"
	"fmt"
	"net/url"

	"github.com/grafov/m3u8"

	"github.com/shaunschembri/restreamer/pkg/restream/request"
)

type Playlist struct {
	playlist     m3u8.Playlist
	ListType     m3u8.ListType
	referenceURL *url.URL
}

func (p Playlist) Type() m3u8.ListType {
	return p.ListType
}

func GetPlaylist(ctx context.Context, request request.Request, playlistURL string) (*Playlist, error) {
	response, err := request.Do(ctx, playlistURL)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer response.Body.Close()

	pl, listType, err := m3u8.DecodeFrom(response.Body, true)
	if err != nil {
		return nil, fmt.Errorf("failed to decode playlist: %w", err)
	}

	return &Playlist{
		playlist:     pl,
		ListType:     listType,
		referenceURL: response.Request.URL,
	}, nil
}
