package server

import (
	"context"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/paynejacob/hotcereal/pkg/provider"
	"github.com/paynejacob/hotcereal/pkg/store"
	"github.com/paynejacob/speakerbob/pkg/auth"
	"github.com/paynejacob/speakerbob/pkg/play"
	"github.com/paynejacob/speakerbob/pkg/service"
	"github.com/paynejacob/speakerbob/pkg/sound"
	"github.com/paynejacob/speakerbob/pkg/static"
	"github.com/paynejacob/speakerbob/pkg/websocket"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type Config struct {
	Host          string
	Port          int
	DurationLimit time.Duration
	AuthProviders []auth.Provider
}

type Server struct {
	httpServer     http.Server
	providers      []provider.Provider
	serviceManager service.Manager
}

func NewServer(_store store.Store, config Config) *Server {
	var svr Server

	// Providers
	tokenProvider := auth.TokenProvider{Store: _store}
	userProvider := auth.UserProvider{Store: _store}
	soundProvider := sound.SoundProvider{Store: _store}
	groupProvider := sound.GroupProvider{Store: _store}
	svr.providers = []provider.Provider{&tokenProvider, &userProvider, &soundProvider, &groupProvider}

	router := mux.NewRouter()
	authRouter := router.PathPrefix("/auth").Subrouter()
	apiRouter := router.PathPrefix("/api").Subrouter()

	// Services
	websocketService := &websocket.Service{}
	svr.serviceManager.RegisterService(apiRouter, websocketService)
	svr.serviceManager.RegisterService(apiRouter, &play.Service{
		SoundProvider:    &soundProvider,
		GroupProvider:    &groupProvider,
		WebsocketService: websocketService,
		MaxSoundDuration: config.DurationLimit,
	})
	svr.serviceManager.RegisterService(apiRouter, &sound.Service{
		SoundProvider:    &soundProvider,
		GroupProvider:    &groupProvider,
		WebsocketService: websocketService,
		MaxSoundDuration: config.DurationLimit,
	})
	authService := &auth.Service{
		TokenProvider: &tokenProvider,
		UserProvider:  &userProvider,
		Providers:     config.AuthProviders,
	}
	svr.serviceManager.RegisterService(authRouter, authService)

	router.NotFoundHandler = static.Service{}

	// middleware
	router.Use(handlers.RecoveryHandler())
	router.Use(handlers.ProxyHeaders)
	router.Use(handlers.CompressHandler)
	apiRouter.Use(authService.Handler)

	svr.httpServer.Addr = fmt.Sprintf("%s:%d", config.Host, config.Port)
	svr.httpServer.Handler = router
	svr.httpServer.ReadTimeout = 5 * time.Second
	svr.httpServer.ReadHeaderTimeout = 2 * time.Second
	svr.httpServer.WriteTimeout = 10 * time.Second
	svr.httpServer.IdleTimeout = 60 * time.Second

	return &svr
}

func (s *Server) Run(ctx context.Context) error {
	logrus.Info("Initializing providers")
	for _, p := range s.providers {
		if err := p.Initialize(); err != nil {
			logrus.Errorf("Error initalizing provider: %s", err.Error())
			return err
		}
	}

	logrus.Info("Starting services")
	go s.serviceManager.Run(ctx)

	logrus.Infof("Listening on http://%s", s.httpServer.Addr)

	// watch the context and shutdown the server if it is canceled
	go func() {
		<-ctx.Done()

		logrus.Info("Shutting down HTTP server")
		ctxShutDown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer func() {
			cancel()
		}()

		if err := s.httpServer.Shutdown(ctxShutDown); err != nil && err != http.ErrServerClosed {
			logrus.Errorf("error shutting down HTTP server: %s", err.Error())
		}
	}()

	// we treat the http server as our "main thread" and exit if it exits
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logrus.Error("HTTP server error: %s", err.Error())
		return err
	}

	return nil
}
