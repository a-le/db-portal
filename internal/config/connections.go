package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// ConnectionsConfig holds multiple connection configurations
type ConnectionsConfig map[string]Connection

// Connection represents a single database connection configuration
type Connection struct {
	DBType string `yaml:"db-type"`
	DSN    string `yaml:"dsn"`
}

// UnmarshalYAML is a custom unmarshal function for the Connection struct.
// It loads the DSN from the environment variable if `env_dsn` is specified in the YAML file and set in the environment.
func (c *Connection) UnmarshalYAML(node *yaml.Node) error {
	// Define a separate struct matching the YAML structure including the optional env_dsn field
	type rawConnection struct {
		DBType string `yaml:"db-type"`
		DSN    string `yaml:"dsn"`
		EnvDSN string `yaml:"env-dsn"`
	}

	// Decode into the rawConnection struct first
	var raw rawConnection
	if err := node.Decode(&raw); err != nil {
		return err
	}

	// Set the DBType and DSN from the YAML
	c.DBType = raw.DBType
	c.DSN = raw.DSN

	// If EnvDSN is specified and the corresponding environment variable is set, override the DSN
	if raw.EnvDSN != "" {
		if envDSNValue := os.Getenv(raw.EnvDSN); envDSNValue != "" {
			c.DSN = envDSNValue // Override DSN with environment variable
		}
	}

	return nil
}

/*
// Struct for connections.yaml
type ConnectionsConfig map[string]struct {
	DBType string `yaml:"dbtype"`
	DSN    string `yaml:"dsn"`
	EnvDSN string `yaml:"env_dsn"`
}
*/
