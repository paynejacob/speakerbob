package cmd

import (
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/gorilla/mux"
	"github.com/paynejacob/speakerbob/pkg/play"
	"github.com/paynejacob/speakerbob/pkg/sound"
	"github.com/paynejacob/speakerbob/pkg/static"
	"github.com/paynejacob/speakerbob/pkg/store/badgerdb"
	"github.com/paynejacob/speakerbob/pkg/websocket"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"time"
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

	level, err := logrus.ParseLevel(logLevelString)
	if err != nil {
		panic(err)
	}
	logrus.SetLevel(level)

	durationLimit, err := time.ParseDuration(durationLimitString)
	if err != nil {
		logrus.Fatal(err)
	}

	db, err := badger.Open(badger.DefaultOptions(dataPath))
	if err != nil {
		logrus.Fatal(err)
	}

	_store := badgerdb.Store{
		db,
	}

	soundProvider := sound.NewSoundProvider(_store)
	if err = soundProvider.Initialize(); err != nil {
		logrus.Fatal(err)
	}

	groupProvider := sound.NewGroupProvider(_store)
	if err = groupProvider.Initialize(); err != nil {
		logrus.Fatal(err)
	}

	websocketService := websocket.NewService()
	websocketService.RegisterRoutes(r, "/ws")

	playService := play.NewService(soundProvider, groupProvider, websocketService, durationLimit)
	playService.RegisterRoutes(r, "/play")

	soundService := sound.NewService(soundProvider, groupProvider, websocketService, durationLimit)
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
