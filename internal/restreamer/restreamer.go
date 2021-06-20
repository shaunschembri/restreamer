package restreamer

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/viper"

	"github.com/shaunschembri/restreamer/pkg/restream"
)

const mbMultiplier = 1048576

func start(ctx context.Context, writer io.Writer, streamID string) error {
	streamer := restream.Restream{
		Writer:         writer,
		MaxBandwidth:   uint32(viper.GetFloat64("max-bandwidth") * mbMultiplier),
		ReadBufferSize: int(viper.GetFloat64("read-buffer") * mbMultiplier),
	}

	streams := viper.GetStringMapString("streams")
	streamURL, ok := streams[streamID]
	if !ok {
		return fmt.Errorf("url for stream with id %s not found in config", streamID)
	}

	if err := streamer.Start(ctx, streamURL); err != nil {
		return fmt.Errorf("restreamer error %w", err)
	}

	return nil
}
