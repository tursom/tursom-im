package config

import "net/http"

type Config struct {
	Admin  AdminConfig  `yaml:"admin"`
	Server ServerConfig `yaml:"server"`
}

type AdminConfig struct {
	Id       string `yaml:"id"`
	Password string `yaml:"password"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

func NewConfig() *Config {
	return &Config{
		Admin:  AdminConfig{},
		Server: ServerConfig{Port: 12345},
	}
}

// CheckAdmin
// @return appId null if failed
func (c *AdminConfig) CheckAdmin(r *http.Request) *string {
	appId := r.Header["AppId"][0]
	appToken := r.Header["AppToken"][0]
	if appId != c.Id || appToken != c.Password {
		return nil
	}
	return &appId
}
