package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"log"
	"net/http"
	"speakerbob/internal"
	"speakerbob/internal/models"
	"speakerbob/internal/services"
)

func main() {
	router := mux.NewRouter()
	n := negroni.New(negroni.NewRecovery())

	// Register Logger
	logger := negroni.NewLogger()
	logger.SetFormat(internal.GetConfig().LogFormat)
	n.Use(logger)

	// Configure Server
	registerRoutes(router)
	n.UseHandler(router)

	// Migrate Database
	internal.GetDB().AutoMigrate(&models.Sound{}, &models.Macro{}, &models.PositionalSound{})

	// Start Websocket service
	go services.WSMessageConsumer()

	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", internal.GetConfig().Host, internal.GetConfig().Port), n))
}

// Registers handlers to routes
func registerRoutes(router *mux.Router) {
	// API Routes
	router.HandleFunc("/api/ws", services.GetWSConnect).Methods("GET")
	router.HandleFunc("/api/speak", services.GetSpeak).Methods("GET")

	// Generic Routes
	router.HandleFunc("/status", services.GetStatus).Methods("GET")
	router.Handle("/", http.FileServer(http.Dir("../assets")))
}
