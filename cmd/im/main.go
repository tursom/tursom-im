package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tursom/GoCollections/exceptions"
	"gopkg.in/yaml.v2"

	"github.com/tursom-im/config"
	ctx "github.com/tursom-im/context"
	_ "github.com/tursom-im/handler/handler"
	"github.com/tursom-im/handler/transport/web"
)

func systemInit() *config.Config {
	rand.Seed(time.Now().UnixNano())

	configPath := "config.yaml"

	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--config":
		case "-c":
			i++
			configPath = os.Args[i]
		}
	}

	cfg := config.NewConfig()
	configFile, err := os.ReadFile(configPath)
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

	globalContext := ctx.NewGlobalContext(cfg)
	if globalContext == nil {
		panic(exceptions.NewRuntimeException("", nil))
	}

	webSocketHandler := web.NewWebSocketHandler(globalContext)
	tokenHandler := web.NewTokenHandler(globalContext)

	router := httprouter.New()
	tokenHandler.InitWebHandler(router)
	webSocketHandler.InitWebHandler(router)

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
