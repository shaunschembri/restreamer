package restream

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/grafov/m3u8"

	"github.com/shaunschembri/restreamer/pkg/restream/provider"
	"github.com/shaunschembri/restreamer/pkg/restream/provider/hls"
	"github.com/shaunschembri/restreamer/pkg/restream/request"
)

func (r Restream) Start(ctx context.Context, playlistURL string) error {
	if err := r.init(ctx, playlistURL); err != nil {
		return err
	}

	segmentsContext, cancel := context.WithCancel(ctx)
	defer cancel()
	go r.getSegments(segmentsContext)

	for {
		segments, sleepTime, err := r.SegmentProvider.Get(ctx, r.currentBandwidth)
		if err != nil {
			return fmt.Errorf("failed to get new segments: %w", err)
		}

		for _, segment := range segments {
			r.segments <- segment
		}

		select {
		case <-ctx.Done():
			r.displayStats()
			return nil
		case err := <-r.errors:
			return err
		case <-time.After(sleepTime):
			r.displayStats()
		}
	}
}

func (r Restream) displayStats() {
	statsString := fmt.Sprintf("Streamed: %5.1fMB | Calculated Bandwidth: %4.1fMb/s",
		float64(r.streamedBytes)/mbDivider, float64(r.currentBandwidth)/mbDivider)

	if r.decrypter != nil {
		statsString += fmt.Sprintf(" | Decrypter %s", r.decrypter.info())
	}

	log.Printf("%s | Playlist Type: %s", statsString, r.SegmentProvider.Info())
}

func (r *Restream) detectStream(ctx context.Context, playlistURL, userAgent string, maxBandwidth uint32) (provider.Provider, error) {
	request := request.New(userAgent)
	playlist, err := hls.GetPlaylist(ctx, request, playlistURL)
	if err != nil {
		return nil, fmt.Errorf("cannot get playlist: %w", err)
	}

	switch playlist.Type() {
	case m3u8.MEDIA:
		return hls.NewMedia(request).WithPlaylistURL(playlistURL), nil
	case m3u8.MASTER:
		return hls.NewMaster(request, maxBandwidth).WithPlaylist(playlist), nil
	default:
		return nil, fmt.Errorf("invalid playlist list type found at %s", playlistURL)
	}
}
