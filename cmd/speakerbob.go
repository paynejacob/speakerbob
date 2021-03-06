package main

import "C"
import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/urfave/cli"
	"github.com/urfave/negroni"
	"log"
	"net/http"
	"os"
	"speakerbob/internal"
	"speakerbob/internal/api"
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
				var user = api.NewUser(c.Args().Get(0), c.Args().Get(1), c.Args().Get(0))
				config := internal.GetConfig()
				db := internal.GetDB(config.DBURL)
				if err := db.Create(&user).Error; err != nil {
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
	log.Print("starting Speakerbob")

	config := internal.GetConfig()
	db := internal.GetDB(config.DBURL)
	router := mux.NewRouter()
	n := negroni.New(negroni.NewRecovery())
	logger := negroni.NewLogger()

	// configure logger
	logger.SetFormat(config.LogFormat)
	n.Use(logger)

	log.Println("creating services")
	wsService := api.NewWebsocketService(config.MessageBrokerURL, db)
	searchService := api.NewSearchService(config.SearchBackendURL)
	authService := api.NewAuthenticationService(config.AuthBackendURL, config.CookieName, config.TokenTTL, config.PageSize, db)
	soundService := api.NewSoundService(config.SoundBackendURL, config.PageSize, config.MaxSoundLength, db, wsService, searchService, config.BluemixAPIKey)

	log.Print("registering routes")
	authService.RegisterRoutes(router, "/auth")
	searchService.RegisterRoutes(router, "/api/search").Use(authService.AuthenticationMiddleware)
	soundService.RegisterRoutes(router, "/api").Use(authService.AuthenticationMiddleware)
	wsService.RegisterRoutes(router, "/api/ws").Use(authService.AuthenticationMiddleware)
	router.Handle("/", http.FileServer(http.Dir("/etc/speakerbob/assets")))

	log.Print("Hydrating search index")
	hydrateIndex(searchService, db)

	log.Print("starting ws consumer")
	go wsService.WSMessageConsumer()

	log.Print("starting web server")
	log.Printf("listening on %s:%v", config.Host, config.Port)
	n.UseHandler(router)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", config.Host, config.Port), n))
}

func hydrateIndex(searchService *api.SearchService, db *gorm.DB) {
	sounds, _ := db.Model(api.Sound{}).Rows()
	defer func() {_ = sounds.Close() }()
	for sounds.Next() {
		var sound api.Sound

		_ = db.ScanRows(sounds, &sound)
		_ = searchService.UpdateResult(sound)

	}

	macros, _ := db.Model(api.Macro{}).Rows()
	defer func() {_ = macros.Close() }()
	for macros.Next() {
		var macro api.Macro

		_ = db.ScanRows(macros, &macro)
		_ = searchService.UpdateResult(macro)

	}
}
