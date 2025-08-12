package handlers

import (
	"db-portal/internal/config"
	"db-portal/internal/internaldb"
	"time"
)

type Services struct {
	Store           *internaldb.Store
	CommandsConfig  *config.Config[config.CommandsConfig]
	ServerConfig    *config.Config[config.Server]
	clockResolution time.Duration
}
