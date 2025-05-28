package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"db-portal/internal/auth"
	"db-portal/internal/config"
	"db-portal/internal/export"
	"db-portal/internal/handlers"
	"db-portal/internal/internaldb"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {

	// get config folder path
	configPath, err := config.NewConfigPath(os.Args)
	if err != nil {
		log.Fatalf("error getting config path: %s", err)
	}

	// initialize internal DB store
	store, err := internaldb.NewStore(configPath)
	if err != nil {
		log.Fatalf("error initializing connections: %v", err)
	}
	fmt.Printf("DB file %v will be used as internal DB\n", store.DBPath)

	// warm up internal DB in the background so that 1st request is not slow
	fmt.Println("DB warmup start")
	go func() {
		if err := store.WarmUp(); err != nil {
			log.Printf("error warming up internal DB: %v", err)
		} else {
			fmt.Println("DB warmup done")
		}
	}()

	// load server config file
	serverConfig := config.New[config.Server](configPath + "/server.yaml")
	if err := serverConfig.Load(); err != nil {
		log.Fatalf("error loading %s file: %s", serverConfig.Filename, err)
	}

	// load sql commands config file
	commandsConfig := config.New[config.CommandsConfig](configPath + "/commands.yaml")
	if err := commandsConfig.Load(); err != nil {
		log.Fatalf("error loading %s file: %s", commandsConfig.Filename, err)
	}

	// gen a random JWT secret key
	jwtSecretKey := auth.RandomString(32)

	// initialize services for handlers
	svcs := &handlers.Services{
		Store:          store,
		CommandsConfig: &commandsConfig,
		ServerConfig:   &serverConfig,
		Exporter:       &export.DefaultExporter{},
		JWTSecretKey:   jwtSecretKey,
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Compress(5, "text/html", "text/css", "application/json", "text/javascript"))

	authMw := auth.Auth(store, jwtSecretKey) // auth middleware
	r.Route("/api", func(api chi.Router) {
		api.Use(authMw)

		api.Get("/connect/{conn}", svcs.ConnectHandler)
		api.Post("/export", svcs.ExportHandler)
		api.Post("/query", svcs.QueryHandler)
		api.Get("/command/{conn}/{schema}/{command}", svcs.CommandHandler)
		api.Get("/config/cnxnames", svcs.CnxNamesHandler)
		api.Get("/clockresolution", svcs.ClockResolutionHandler)
	})
	r.With(authMw).Get("/", svcs.IndexHandler)
	r.With(authMw).Get("/logout", svcs.LogoutHandler)
	r.Get("/hash/{string}", svcs.HashHandler)
	r.Handle("/web/*", svcs.StaticFileHandler())

	// create HTTP server
	httpServer := &http.Server{
		Addr:    serverConfig.Data.Addr,
		Handler: r,
	}

	// start the server with HTTPS if cert and key files are provided, otherwise use HTTP
	scd := serverConfig.Data
	useHTTPS := scd.CertFile != "" && scd.KeyFile != ""
	if useHTTPS {
		fmt.Printf("server is listening at https://%s\n", httpServer.Addr)
		if err := httpServer.ListenAndServeTLS(scd.CertFile, scd.KeyFile); err != nil {
			log.Fatalf("server failed to start at https://%s: %v", httpServer.Addr, err)
		}
	} else {
		fmt.Printf("server is listening at http://%s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatalf("server failed to start at https://%s: %v", httpServer.Addr, err)
		}
	}
}
