package config

import (
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"
	"gopkg.in/yaml.v2"
	"net/http"
	"strings"
)

type (
	Config struct {
		lang.BaseObject `yaml:"-" json:"-"`
		Admin           AdminConfig  `yaml:"admin" json:"admin"`
		Server          ServerConfig `yaml:"server" json:"server"`
		SSL             SSL          `yaml:"ssl" json:"ssl"`
		Node            NodeConfig   `yaml:"node" json:"node"`
	}

	AdminConfig struct {
		lang.BaseObject `yaml:"-" json:"-"`
		Id              string `yaml:"id" json:"id"`
		Password        string `yaml:"password" json:"password"`
	}

	ServerConfig struct {
		lang.BaseObject `yaml:"-" json:"-"`
		Port            int `yaml:"port" json:"port"`
	}

	SSL struct {
		lang.BaseObject `yaml:"-" json:"-"`
		Cert            string `yaml:"cert" json:"cert"`
		Key             string `yaml:"key" json:"key"`
		Enable          bool   `yaml:"enable" json:"enable"`
	}

	NodeConfig struct {
		lang.BaseObject `yaml:"-" json:"-"`
		NodeMax         int32 `yaml:"nodeMax" json:"nodeMax"`
	}
)

func NewConfig() *Config {
	return &Config{
		Admin:  AdminConfig{},
		Server: ServerConfig{Port: 12345},
		SSL:    SSL{Enable: false},
		Node:   NodeConfig{NodeMax: 4096},
	}
}

func (c *Config) String() string {
	sw := &strings.Builder{}
	encoder := yaml.NewEncoder(sw)
	if err := encoder.Encode(c); err != nil {
		panic(exceptions.Package(err))
	}
	return sw.String()
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
