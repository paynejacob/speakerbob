package health

import (
	"context"
	"github.com/gorilla/mux"
	"net/http"
)

type Service struct{}

func (s Service) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/healthz", s.healthz)
}

func (s Service) Run(context.Context) {}

func (s Service) healthz(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("ok"))
}
