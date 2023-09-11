package config

type Config struct {
}

var (
	Build   string
	Version string
)

func New() *Config {
	return &Config{}
}
