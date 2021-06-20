package restream

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/grafov/m3u8"
)

type hlsMaster struct {
	hlsMedia
	resolution       string
	maxBandwidth     uint32
	variantBandwidth uint32
	playlist         *m3u8.MasterPlaylist
	referenceURL     *url.URL
}

func (h hlsMaster) Info() string {
	infoStr := fmt.Sprintf("Master | Bandwidth: %3.1fMb/s", float32(h.variantBandwidth)/mbDivider)
	if h.resolution != "" {
		infoStr += fmt.Sprintf(" | Resolution: %s", h.resolution)
	}

	return infoStr
}

func (h *hlsMaster) Get(ctx context.Context, bandwidth uint32) ([]segment, time.Duration, error) {
	if err := h.selectVariant(bandwidth); err != nil {
		return nil, 0, err
	}

	return h.hlsMedia.Get(ctx, bandwidth)
}

func (h *hlsMaster) selectVariant(streamSpeed uint32) error {
	var targetVariant *m3u8.Variant
	minDiff := streamSpeed

	for _, variant := range h.playlist.Variants {
		if variant.Bandwidth > h.maxBandwidth || variant.Bandwidth > streamSpeed {
			continue
		}

		diff := streamSpeed - variant.Bandwidth
		if diff >= minDiff && targetVariant != nil {
			continue
		}

		minDiff = diff
		targetVariant = variant
	}

	parsedURI, err := h.request.resolveReference(targetVariant.URI, h.referenceURL)
	if err != nil {
		return err
	}

	h.mediaPlaylistURL = parsedURI.String()
	h.resolution = targetVariant.Resolution
	h.variantBandwidth = targetVariant.Bandwidth

	return nil
}
