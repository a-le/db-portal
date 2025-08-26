package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"db-portal/internal/config"
	"db-portal/internal/handlers"
	"db-portal/internal/internaldb"
	"db-portal/internal/meta"
	"db-portal/internal/security"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {

	// Declare and parse command line flags
	configPathFlag := flag.String("config", "", "Path to config folder")
	setMasterPasswordFlag := flag.String("set-master-password", "", "Set new password for master user (id=1)")
	flag.Parse()

	configPath, err := config.NewConfigPath(*configPathFlag)
	if err != nil {
		log.Fatalf("error getting config path: %s", err)
	}

	dataPath, err := config.NewConfigPath(*configPathFlag)
	if err != nil {
		log.Fatalf("error getting config path: %s", err)
	}

	// Initialize internal DB store
	store, err := internaldb.NewStore(dataPath)
	if err != nil {
		log.Fatalf("error initializing connections: %v", err)
	}
	fmt.Printf("DB file %v will be used as internal DB\n", store.DBPath)

	// Handle set-master-password flag
	if *setMasterPasswordFlag != "" {
		if err := store.SetMasterUserPassword(*setMasterPasswordFlag); err != nil {
			log.Fatalf("failed to set master password: %v", err)
		}
		fmt.Println("Master user password updated successfully")
	}

	// Load server config file
	path := filepath.Join(configPath, "server.yaml")
	serverConfig := config.New[config.Server](path)
	if err := serverConfig.Load(); err != nil {
		log.Fatalf("error loading %s file: %s", serverConfig.Filename, err)
	}
	useHTTPS := serverConfig.Data.CertFile != "" && serverConfig.Data.KeyFile != ""

	// Load sql commands config file
	path = filepath.Join(configPath, "commands.yaml")
	commandsConfig := config.New[config.CommandsConfig](path)
	if err := commandsConfig.Load(); err != nil {
		log.Fatalf("error loading %s file: %s", commandsConfig.Filename, err)
	}

	// JWTSecretKey is read from file, it is generated if file not exists
	path = filepath.Join(configPath, meta.JWTKeyFileName)
	key, err := security.LoadJWTSecretKey(path)
	if err != nil {
		key, err = security.GenerateJWTSecretKey()
		if err != nil {
			log.Fatalf("error generating JWT secret key: %s", err)
		} else {
			fmt.Println("JWT secret key generated successfully")
		}
		if err := security.SaveJWTSecretKey(path, key); err != nil {
			log.Fatalf("error saving file: %s", err)
		} else {
			fmt.Println("JWT secret key file saved successfully")
		}
	}
	security.JWTSecretKey = key

	// Initialize services for handlers
	svcs := &handlers.Services{
		Store:          store,
		CommandsConfig: &commandsConfig,
		ServerConfig:   &serverConfig,
	}

	r := chi.NewRouter()

	// Core middleware stack
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.Compress(5, "text/html", "text/css", "application/json", "text/javascript"))

	// Public routes
	r.Get("/", svcs.IndexHandler)
	r.Handle("/web/*", svcs.StaticFileHandler())
	r.Post("/api/auth/login", svcs.HandleLogin)

	// Private routes
	r.Route("/api", func(api chi.Router) {
		api.Use(security.JWTMiddleware) // apply JWT auth middleware

		api.Get("/users", svcs.HandleListUsers)
		api.Post("/users", svcs.HandleCreateUser)

		api.Get("/users/{username}/data-sources", svcs.HandleListUserDataSources)
		api.Get("/users/{username}/available-data-sources", svcs.HandleListUserAvailableDataSources)
		api.Get("/users/{username}/data-sources/{dsName}/test", svcs.HandleUserDataSourceTest)
		api.Post("/users/{username}/data-sources/{dsName}", svcs.HandleCreateUserDataSource)
		api.Delete("/users/{username}/data-sources/{dsName}", svcs.HandleDeleteUserDataSource)

		api.Post("/data-sources/test", svcs.HandleDataSourceTest) // do not use GET, use POST to receive DSN location
		api.Post("/data-sources", svcs.HandleCreateDataSource)

		api.Get("/vendors", svcs.HandleListVendors)

		api.Get("/clock-resolution", svcs.HandleClockResolution)

		api.Get("/command/{dsName}/{command}", svcs.CommandHandler)
		api.Get("/command/{dsName}/{schema}/{command}", svcs.CommandHandler)

		api.Post("/query/{dsName}", svcs.QueryHandler)
		api.Post("/query/{dsName}/{schema}", svcs.QueryHandler)

		api.Post("/copy", svcs.CopyHandler)
	})

	// Create HTTP server
	httpServer := &http.Server{
		Addr:    serverConfig.Data.Addr,
		Handler: r,
	}

	// Start the server with HTTPS if cert and key files are provided, otherwise use HTTP
	if useHTTPS {
		fmt.Printf("server is listening at https://%s\n", httpServer.Addr)
		if err := httpServer.ListenAndServeTLS(serverConfig.Data.CertFile, serverConfig.Data.KeyFile); err != nil {
			log.Fatalf("server failed to start at https://%s: %v", httpServer.Addr, err)
		}
	} else {
		fmt.Printf("server is listening at http://%s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatalf("server failed to start at http://%s: %v", httpServer.Addr, err)
		}
	}
}
