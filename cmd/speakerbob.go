package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/urfave/cli"
	"github.com/urfave/negroni"
	"log"
	"net/http"
	"os"
	"speakerbob/internal"
	"speakerbob/internal/authentication"
	"speakerbob/internal/services"
	"speakerbob/internal/sound"
)

func main() {
	app := cli.NewApp()
	app.Name = "Speakerbob"
	app.Usage = "A distributed soundboard."
	app.Action = func(c *cli.Context) error {
		serve()
		return nil
	}
	app.Commands = []cli.Command{
		{
			Name:  "serve",
			Usage: "Run the Speakerbob service.",
			Action: func(c *cli.Context) error {
				serve()
				return nil
			},
		},
		{
			Name:  "adduser",
			Usage: "Create a new user.",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "username", Usage: "the new user's username"},
				cli.StringFlag{Name: "password", Usage: "the new user's password"},
			},
			Action: func(c *cli.Context) error {
				var user = authentication.NewUser(c.Args().Get(0), c.Args().Get(1), c.Args().Get(0))
				if err := internal.GetDB().Create(&user).Error; err != nil {
					log.Printf("An error occured creating the user: %v", err)
					return nil
				}

				log.Printf("User \"%s\" sucessfully created!", c.Args().Get(0))
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func serve() {
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
	internal.GetDB().AutoMigrate(&sound.Sound{}, &sound.Macro{}, &sound.PositionalSound{}, &authentication.User{})

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
	router.HandleFunc("/api/ws", services.WSConnect).Methods("GET")
	router.HandleFunc("/api/speak", services.Speak).Methods("GET")

	router.HandleFunc("/api/login", authentication.Login).Methods("POST")
	router.HandleFunc("/api/logout", authentication.Logout).Methods("GET")

	router.HandleFunc("/api/sound", sound.ListSound).Methods("GET")
	router.HandleFunc("/api/sound", sound.CreateSound).Methods("POST")
	router.HandleFunc("/api/sound/{id}", sound.GetSound).Methods("GET")
	router.HandleFunc("/api/sound/{id}/download", sound.DownloadSound).Methods("GET")

	router.HandleFunc("/api/macro", sound.ListMacro).Methods("GET")
	router.HandleFunc("/api/macro", sound.CreateMacro).Methods("POST")
	router.HandleFunc("/api/macro/{id}", sound.GetMacro).Methods("GET")
	router.HandleFunc("/api/macro/{id}/download", sound.DownloadMacro).Methods("GET")

	// Generic Routes
	router.HandleFunc("/status", services.Status).Methods("GET")
	router.Handle("/", http.FileServer(http.Dir("../assets")))
}
