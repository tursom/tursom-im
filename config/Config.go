package config

import "net/http"

type Config struct {
	Admin  AdminConfig  `yaml:"admin"`
	Server ServerConfig `yaml:"server"`
	SSL    SSL          `yaml:"ssl"`
	Node   NodeConfig   `yaml:"node"`
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

type NodeConfig struct {
	NodeMax int32 `yaml:"nodeMax"`
}

func NewConfig() *Config {
	return &Config{
		Admin:  AdminConfig{},
		Server: ServerConfig{Port: 12345},
		SSL:    SSL{Enable: false},
		Node:   NodeConfig{NodeMax: 4096},
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
