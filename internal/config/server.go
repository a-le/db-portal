package config

// Struct for server.yaml
type Server struct {
	Addr               string `yaml:"addr"`
	HtpasswdFile       string `yaml:"htpasswd-file"`
	DbTimeout          int    `yaml:"db-timeout"`
	MaxResultsetLength int    `yaml:"max-resultset-length"`
	// JWTSecretKey       string `yaml:"jwt-secret-key"`
	CertFile string `yaml:"cert-file"`
	KeyFile  string `yaml:"key-file"`
}

// // UnmarshalYAML is a custom unmarshal function for the Server struct.
// // It loads the JwtSecretKey from the environment variable if `env-jwt-secret-key` is specified in the YAML file and set in the environment.
// func (s *Server) UnmarshalYAML(node *yaml.Node) error {
// 	// Create a temporary struct to unmarshal the YAML data
// 	type rawServer struct {
// 		Addr               string `yaml:"addr"`
// 		HtpasswdFile       string `yaml:"htpasswd-file"`
// 		DbTimeout          int    `yaml:"db-timeout"`
// 		MaxResultsetLength int    `yaml:"max-resultset-length"`
// 		JWTSecretKey       string `yaml:"jwt-secret-key"`
// 		EnvJwtSecretKey    string `yaml:"env-jwt-secret-key"`
// 		CertFile           string `yaml:"cert-file"`
// 		KeyFile            string `yaml:"key-file"`
// 	}

// 	// Decode the YAML node into the temporary struct
// 	var raw rawServer
// 	if err := node.Decode(&raw); err != nil {
// 		return err
// 	}

// 	// Set the values into the Server struct
// 	s.Addr = raw.Addr
// 	s.HtpasswdFile = raw.HtpasswdFile
// 	s.DbTimeout = raw.DbTimeout
// 	s.MaxResultsetLength = raw.MaxResultsetLength
// 	s.CertFile = raw.CertFile
// 	s.KeyFile = raw.KeyFile

// 	// Check if the environment variable `env-jwt-secret-key` is provided
// 	if raw.EnvJwtSecretKey != "" {
// 		// Attempt to get the value from the environment variable
// 		if envValue := os.Getenv(raw.EnvJwtSecretKey); envValue != "" {
// 			s.JWTSecretKey = envValue // Set the JWTSecretKey from the environment variable
// 		} else {
// 			// If the environment variable is not set, keep the value from the YAML (or leave it empty if not set)
// 			s.JWTSecretKey = raw.JWTSecretKey
// 		}
// 	} else {
// 		// If no env key is specified, fall back to the value in the YAML for JWTSecretKey
// 		s.JWTSecretKey = raw.JWTSecretKey
// 	}

// 	return nil
// }
