package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tursom/GoCollections/exceptions"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
	"tursom-im/config"
	"tursom-im/context"
	"tursom-im/handler"
	"tursom-im/utils"
)

func SystemInit() (*config.Config, error) {
	rand.Seed(time.Now().UnixNano())
	utils.InitWatchDog(time.Second * 60)

	cfg := config.NewConfig()
	configFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return nil, exceptions.Package(err)
	}
	//fmt.Println(string(configFile))
	err = yaml.Unmarshal(configFile, cfg)
	return cfg, exceptions.Package(err)
}

func main() {
	cfg, err := SystemInit()
	if err != nil {
		exceptions.Print(err)
		return
	}
	fmt.Println(cfg)

	globalContext := context.NewGlobalContext(cfg)
	if globalContext == nil {
		os.Exit(-1)
	}

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
		exceptions.Print(err)
		return
	}
}
