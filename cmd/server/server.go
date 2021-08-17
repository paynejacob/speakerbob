// +build !windows

package server

import (
	"context"
	"github.com/dgraph-io/badger/v3"
	"github.com/paynejacob/speakerbob/pkg/server"
	"github.com/paynejacob/speakerbob/pkg/store/badgerdb"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/signal"
)

var configPath string

func init() {
	Command.PersistentFlags().StringVar(&configPath, "config", "", "Path to a speakerbob server configuration file.")
}

var Command = &cobra.Command{
	Use:   "server",
	Short: "Run the speakerbob server.",
	Long:  `Run the speakerbob server.`,
	Run:   Server,
}

func Server(*cobra.Command, []string) {
	// load configuration
	config, err := parseConfiguration(configPath)
	if err != nil {
		logrus.Fatal(err)
	}

	// setup the store
	badgerdbOptions := badger.DefaultOptions(config.DataPath)
	badgerdbOptions.Logger = logrus.StandardLogger()
	db, err := badger.Open(badgerdbOptions)
	if err != nil {
		logrus.Fatal(err)
	}

	_store := badgerdb.Store{
		DB: db,
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())

	// watch for shutdown signals
	go func() {
		oscall := <-c
		log.Printf("system call:%+v", oscall)
		cancel()
	}()

	logrus.Info("Starting Speakerbob server")
	s := server.NewServer(_store, server.Config{
		Host:          config.Host,
		Port:          config.Port,
		DurationLimit: config.DurationLimit,
		AuthProviders: config.Providers(),
	})
	if err = s.Run(ctx); err != nil {
		logrus.Errorf("server exited unexpectedly: %s", err.Error())
	}

	if err = _store.Close(); err != nil {
		logrus.Errorf("error syncing store: %s", err.Error())
	}
}
