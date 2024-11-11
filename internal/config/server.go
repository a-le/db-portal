package config

// Struct for server.yaml
type Server struct {
	Addr               string `yaml:"addr"`
	HtpasswdPath       string
	DbTimeout          int    `yaml:"db-timeout"`
	MaxResultsetLength int    `yaml:"max-resultset-length"`
	CertFile           string `yaml:"cert-file"`
	KeyFile            string `yaml:"key-file"`
}
