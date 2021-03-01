package cmd

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/paynejacob/speakerbob/pkg/play"
	"github.com/paynejacob/speakerbob/pkg/sound"
	"github.com/paynejacob/speakerbob/pkg/static"
	"github.com/paynejacob/speakerbob/pkg/websocket"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

var (
	host string
	port int
)

func init() {
	hostFlag := "host"
	serverCmd.Flags().StringVar(&host, hostFlag, "0.0.0.0", "Host speakerbob will listen on.")
	_ = viper.BindPFlag(hostFlag, rootCmd.PersistentFlags().Lookup(hostFlag))

	portFlag := "port"
	serverCmd.Flags().IntVar(&port, portFlag, 80, "Port speakerbob will listen on.")
	_ = viper.BindPFlag(portFlag, rootCmd.PersistentFlags().Lookup(portFlag))

	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the speakerbob server.",
	Long:  `Run the speakerbob server.`,
	Run:   Server,
}

func Server(*cobra.Command, []string) {
	r := mux.NewRouter()

	minioClient, err := minio.New(s3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s3Key, s3Secret, ""),
		Secure: true,
	})
	if err != nil {
		logrus.Fatal(err)
	}

	soundStore := sound.NewStore(minioClient, s3Bucket, durationLimit)

	websocketService := websocket.NewService()
	websocketService.RegisterRoutes(r, "/ws")

	playService := play.NewService(soundStore, websocketService)
	playService.RegisterRoutes(r, "/play")

	soundService := sound.NewService(soundStore)
	soundService.RegisterRoutes(r, "/sound")

	staticService := static.NewService()
	staticService.RegisterRoutes(r, "")

	go websocketService.Run()
	go playService.Run()
	go soundService.Run()

	http.Handle("/", r)

	logrus.Infof("Server started listening on http://%s:%d", host, port)

	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil))
}