package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tursom-im/config"
	"github.com/tursom-im/context"
	"github.com/tursom-im/handler"
	"github.com/tursom-im/utils"
	"github.com/tursom/GoCollections/exceptions"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func systemInit() *config.Config {
	rand.Seed(time.Now().UnixNano())
	utils.InitWatchDog()

	cfg := config.NewConfig()
	configFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic(exceptions.Package(err))
	}
	//fmt.Println(string(configFile))
	if err = yaml.Unmarshal(configFile, cfg); err != nil {
		panic(exceptions.Package(err))
	}

	return cfg
}

func main() {
	cfg := systemInit()
	fmt.Println(cfg)

	globalContext := context.NewGlobalContext(cfg)
	if globalContext == nil {
		panic(exceptions.NewRuntimeException(nil, "init GlobalContext failed", nil))
	}

	webSocketHandler := handler.NewWebSocketHandler(globalContext)
	tokenHandler := handler.NewTokenHandler(globalContext)

	router := httprouter.New()
	tokenHandler.InitWebHandler("", router)
	webSocketHandler.InitWebHandler("", router)

	fmt.Println("server will start on port " + strconv.Itoa(cfg.Server.Port))
	runServer(cfg, router)
}

func runServer(cfg *config.Config, router *httprouter.Router) {
	var err error
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
