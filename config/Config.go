package config

import "net/http"

type AdminConfig struct {
	Id       string `yaml:"id"`
	Password string `yaml:"password"`
}

type Config struct {
	Admin AdminConfig `yaml:"admin"`
}

func NewConfig() *Config {
	return &Config{
		Admin: AdminConfig{},
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
