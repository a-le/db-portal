package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"db-portal/internal/auth"
	"db-portal/internal/config"
	"db-portal/internal/handlers"
	"db-portal/internal/internaldb"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	var err error

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
		JWTSecretKey:   jwtSecretKey,
	}

	// HTTP services router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Compress(5, "text/html", "text/css", "application/json", "text/javascript"))

	// connect endpoint
	r.With(auth.Auth(store, jwtSecretKey)).Get("/api/connect/{conn}", svcs.ConnectHandler)

	// export endpoint
	r.With(auth.Auth(store, jwtSecretKey)).Post("/api/export", svcs.ExportHandler)

	// query endpoint
	r.With(auth.Auth(store, jwtSecretKey)).Post("/api/query", svcs.QueryHandler)

	// command endpoint. Those are SQL statement used by the UI
	r.With(auth.Auth(store, jwtSecretKey)).Get("/api/command/{conn}/{schema}/{command}", svcs.CommandHandler)

	// DB connections list
	r.With(auth.Auth(store, jwtSecretKey)).Get("/api/config/cnxnames", svcs.CnxNamesHandler)

	// estimate clock resolution (result is cached)
	r.With(auth.Auth(store, jwtSecretKey)).Get("/api/clockresolution", svcs.ClockResolutionHandler)

	// cnxnames endpoint
	r.With(auth.Auth(store, jwtSecretKey)).Get("/api/config/cnxnames", svcs.CnxNamesHandler)

	// index page
	r.With(auth.Auth(store, jwtSecretKey)).Get("/", svcs.IndexHandler)

	// return a bcrypt hash of a string (useful for password hashing)
	// there is some salt in the hash, so the result will be different each time
	r.Get("/hash/{string}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(auth.HashPassword(chi.URLParam(r, "string"))))
	})

	// logout endpoint. Is meant to be used with bad credentials so that the browser forgets those credentials
	r.With(auth.Auth(store, jwtSecretKey)).Get("/logout", func(w http.ResponseWriter, r *http.Request) {
		// nothing to do
	})

	// serve static files
	r.Handle("/web/*", http.StripPrefix("/web/", http.FileServer(http.Dir("./web"))))

	// create HTTP server
	httpServer := &http.Server{
		Addr:    serverConfig.Data.Addr,
		Handler: r,
	}

	// if the server config has a cert and key file, use HTTPS, otherwise use HTTP
	if serverConfig.Data.CertFile != "" && serverConfig.Data.KeyFile != "" {
		fmt.Printf("HTTPS server is listening on %s\n", serverConfig.Data.Addr)
		// start HTTPS server
		if err := httpServer.ListenAndServeTLS(serverConfig.Data.CertFile, serverConfig.Data.KeyFile); err != nil {
			log.Fatalf("main: HTTPS server failed to start on %s: %v", httpServer.Addr, err)
		}
	} else {
		fmt.Printf("HTTP server is listening on %s\n", serverConfig.Data.Addr)
		// start HTTP server
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatalf("main: HTTP server failed to start on %s: %v", httpServer.Addr, err)
		}
	}

}
