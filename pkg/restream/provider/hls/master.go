package hls

import (
	"context"
	"fmt"
	"time"

	"github.com/grafov/m3u8"

	"github.com/shaunschembri/restreamer/pkg/restream/provider"
	"github.com/shaunschembri/restreamer/pkg/restream/request"
)

type Master struct {
	media            *Media
	playlist         *Playlist
	resolution       string
	maxBandwidth     uint32
	variantBandwidth uint32
}

func NewMaster(request request.Request, maxBandwidth uint32) *Master {
	return &Master{
		media:        NewMedia(request),
		maxBandwidth: maxBandwidth,
	}
}

func (m Master) WithPlaylist(playlist *Playlist) *Master {
	m.playlist = playlist
	return &m
}

func (m Master) Info() string {
	infoStr := fmt.Sprintf("Master | Bandwidth: %3.1fMb/s", float32(m.variantBandwidth)/mbDivider)
	if m.resolution != "" {
		infoStr += fmt.Sprintf(" | Resolution: %s", m.resolution)
	}

	return infoStr
}

func (m *Master) Get(ctx context.Context, bandwidth uint32) ([]provider.Segment, time.Duration, error) {
	if err := m.selectVariant(bandwidth); err != nil {
		return nil, 0, err
	}

	return m.media.Get(ctx, bandwidth)
}

func (m *Master) selectVariant(streamSpeed uint32) error {
	var targetVariant *m3u8.Variant
	minDiff := streamSpeed

	for _, variant := range m.playlist.playlist.(*m3u8.MasterPlaylist).Variants {
		if variant.Bandwidth > m.maxBandwidth || variant.Bandwidth > streamSpeed {
			continue
		}

		diff := streamSpeed - variant.Bandwidth
		if diff >= minDiff && targetVariant != nil {
			continue
		}

		minDiff = diff
		targetVariant = variant
	}

	parsedURI, err := m.media.request.ResolveReference(targetVariant.URI, m.playlist.referenceURL)
	if err != nil {
		return fmt.Errorf("cannot resolve reference: %w", err)
	}

	m.media = m.media.WithPlaylistURL(parsedURI.String())
	m.resolution = targetVariant.Resolution
	m.variantBandwidth = targetVariant.Bandwidth

	return nil
}
