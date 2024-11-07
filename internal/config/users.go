package config

// Struct for users.yaml
type UsersConfig map[string]struct {
	Connections []string `yaml:"connections"`
}
