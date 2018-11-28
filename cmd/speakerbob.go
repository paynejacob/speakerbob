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
	log.Print("Starting Speakerbob")
	router := mux.NewRouter()
	n := negroni.New(negroni.NewRecovery())

	// Register Logger
	logger := negroni.NewLogger()
	logger.SetFormat(internal.GetConfig().LogFormat)
	n.Use(logger)

	log.Print("Registering routes")
	registerRoutes(router)
	n.UseHandler(router)

	log.Print("Migrating database")
	internal.GetDB().AutoMigrate(&models.Sound{}, &models.Macro{}, &models.PositionalSound{})

	log.Printf("Verifying audio bucket")
	err := internal.GetMinioClient().MakeBucket(internal.GetConfig().SoundBucketName, "us-east-1")
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, err := internal.GetMinioClient().BucketExists(internal.GetConfig().SoundBucketName)
		if err == nil && exists {
			log.Print("Audio bucket already exists.")
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Audio bucket was created %s\n", internal.GetConfig().SoundBucketName)
	}

	log.Print("Starting WS Consumer")
	go services.WSMessageConsumer()

	log.Print("Starting Web Server")
	log.Printf("Listening on %s:%v", internal.GetConfig().DBHost, internal.GetConfig().Port)
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
