package restreamer

import (
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start HTTP server",
	Run: func(cmd *cobra.Command, args []string) {
		addr := fmt.Sprintf("%s:%d", viper.GetString("server.address"), viper.GetInt("server.port"))

		http.HandleFunc("/", httpStream)

		log.Printf("Starting HTTP server on %s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatalln(err)
		}
	},
}

func init() {
	serverCmd.Flags().IntP("http-port", "p", 1230, "http server listening port")
	serverCmd.Flags().StringP("http-address", "a", "127.0.0.1", "http server bind address")

	bindFlagToConfig(serverCmd, "http-port", "server.port")
	bindFlagToConfig(serverCmd, "http-address", "server.address")

	rootCmd.AddCommand(serverCmd)
}

func httpStream(writer http.ResponseWriter, request *http.Request) {
	streamID := request.URL.Path[1:]

	log.Printf("Starting to restream stream with id %s", streamID)
	if err := start(request.Context(), writer, streamID); err != nil {
		log.Println(err.Error())
		return
	}

	log.Printf("Restream of stream with id %s stopped", streamID)
}
