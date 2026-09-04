package configuration

type ApplicationConfig struct {
	Server ServerConfig `yaml:"server" mapstructure:"server"`
}

type ServerConfig struct {
	Port           string   `yaml:"port" mapstructure:"port"`
	AllowedOrigins []string `yaml:"allowedOrigins" mapstructure:"allowedOrigins"`
}
