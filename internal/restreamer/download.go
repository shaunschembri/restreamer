package restreamer

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download and save stream",
	Run: func(cmd *cobra.Command, args []string) {
		streamID, _ := cmd.Flags().GetString("stream-id")
		fileName, _ := cmd.Flags().GetString("filename")
		if fileName == "" {
			fileName = filepath.Join(viper.GetString("download.path"), fmt.Sprintf("%s_%d.ts", streamID, time.Now().Unix()))
		}

		file, err := os.Create(fileName)
		if err != nil {
			log.Printf("Error: cannot open file %s", fileName)
			return
		}

		duration, _ := cmd.Flags().GetDuration("duration")
		segmentsContext, cancel := context.WithTimeout(context.Background(), duration)
		defer cancel()

		log.Printf("Starting to download stream with id %s for %v", streamID, duration)
		if err := start(segmentsContext, file, streamID); err != nil {
			log.Printf("Error: %v", err)
		}
		log.Printf("Download stream with id %s stopped", streamID)

		file.Close()
	},
}

func init() {
	downloadCmd.Flags().StringP("download-path", "d", ".", "path to store downloaded media")
	downloadCmd.Flags().StringP("filename", "f", "", "filename of downloaded media")
	downloadCmd.Flags().StringP("stream-id", "s", "", "stream id")
	downloadCmd.Flags().DurationP("duration", "t", time.Hour*12, "stream duration")

	bindFlagToConfig(downloadCmd, "download-path", "download.path")
	if err := downloadCmd.MarkFlagRequired("stream-id"); err != nil {
		log.Fatalln(err)
	}

	rootCmd.AddCommand(downloadCmd)
}
