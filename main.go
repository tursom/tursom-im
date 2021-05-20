package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"
	"tursom-im/config"
	"tursom-im/context"
	"tursom-im/handler"
)

func SystemInit() (*config.Config, error) {
	rand.Seed(time.Now().UnixNano())

	cfg := config.NewConfig()
	configFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}
	//fmt.Println(string(configFile))
	err = yaml.Unmarshal(configFile, cfg)
	return cfg, err
}

func main() {
	cfg, err := SystemInit()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(cfg)

	globalContext := context.NewGlobalContext(cfg)
	webSocketHandler := handler.NewWebSocketHandler(globalContext)
	tokenHandler := handler.NewTokenHandler(globalContext)

	router := httprouter.New()
	tokenHandler.InitWebHandler("", router)
	webSocketHandler.InitWebHandler("", router)

	fmt.Println("server will start on port " + strconv.Itoa(cfg.Server.Port))
	if cfg.SSL.Enable {
		err = http.ListenAndServeTLS(":"+strconv.Itoa(cfg.Server.Port), cfg.SSL.Cert, cfg.SSL.Key, router)
	} else {
		err = http.ListenAndServe(":"+strconv.Itoa(cfg.Server.Port), router)
	}
	if err != nil {
		fmt.Println(err)
		return
	}
}
