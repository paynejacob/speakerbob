package service

import "github.com/gorilla/mux"

type Service interface {
	RegisterRoutes(parent *mux.Router, prefix string)
	Run()
}
