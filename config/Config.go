package config

type AdminConfig struct {
	Id       string `yaml:"id"`
	Password string `yaml:"password"`
	Sig      uint64 `yaml:"sig"`
}

type Config struct {
	Admin AdminConfig `yaml:"admin"`
}

func NewConfig() *Config {
	return &Config{
		Admin: AdminConfig{},
	}
}
