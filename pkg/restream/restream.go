package restream

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/grafov/m3u8"
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

	log.Printf("%s | Playlist Type %s", statsString, r.SegmentProvider.Info())
}

func (r *Restream) getSegmentProvider(ctx context.Context, playlistURL string) error {
	hlsMedia := hlsMedia{
		request: request{
			userAgent: r.UserAgent,
			client:    &http.Client{},
		},
		mediaPlaylistURL: playlistURL,
	}

	playlist, err := hlsMedia.getPlaylist(ctx)
	if err != nil {
		return err
	}

	if playlist.listType == m3u8.MEDIA {
		r.SegmentProvider = &hlsMedia
		return nil
	}

	if playlist.listType == m3u8.MASTER {
		masterPlaylist, ok := playlist.playlist.(*m3u8.MasterPlaylist)
		if !ok {
			return fmt.Errorf("cannot assert to a master playlist")
		}

		hlsMaster := hlsMaster{
			playlist:     masterPlaylist,
			referenceURL: playlist.requestURL,
			hlsMedia:     hlsMedia,
			maxBandwidth: r.MaxBandwidth,
		}

		r.SegmentProvider = &hlsMaster
		return nil
	}

	return errors.New("invalid playlist list type")
}
