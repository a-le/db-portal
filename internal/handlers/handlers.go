package handlers

import (
	"db-portal/internal/config"
	"db-portal/internal/export"
	"db-portal/internal/internaldb"
	"time"
)

type Services struct {
	Store           *internaldb.Store
	CommandsConfig  *config.Config[config.CommandsConfig]
	ServerConfig    *config.Config[config.Server]
	Exporter        export.Exporter
	JWTSecretKey    string
	clockResolution time.Duration
}
