package config

import "net/http"

type Config struct {
	Admin  AdminConfig  `yaml:"admin"`
	Server ServerConfig `yaml:"server"`
	SSL    SSL          `yaml:"ssl"`
}

type AdminConfig struct {
	Id       string `yaml:"id"`
	Password string `yaml:"password"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type SSL struct {
	Cert   string `yaml:"cert"`
	Key    string `yaml:"key"`
	Enable bool   `yaml:"enable"`
}

func NewConfig() *Config {
	return &Config{
		Admin:  AdminConfig{},
		Server: ServerConfig{Port: 12345},
		SSL:    SSL{Enable: false},
	}
}

// CheckAdmin
// @return appId null if failed
func (c *AdminConfig) CheckAdmin(r *http.Request) *string {
	appId := r.Header["App-Id"][0]
	appToken := r.Header["App-Token"][0]
	if appId != c.Id || appToken != c.Password {
		return nil
	}
	return &appId
}
