package main

import (
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"github.com/tursom/GoCollections/exceptions"
	"gopkg.in/yaml.v2"

	"github.com/tursom/tursom-im/config"
	ctx "github.com/tursom/tursom-im/context"
	"github.com/tursom/tursom-im/exception"
	_ "github.com/tursom/tursom-im/handler/request"
	"github.com/tursom/tursom-im/handler/transport/web"
)

func getConfig() *config.Config {
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
	cfg := getConfig()
	log.Info(cfg)

	globalContext := ctx.NewGlobalContext(cfg)
	if globalContext == nil {
		panic(exceptions.NewRuntimeException("", nil))
	}

	webSocketHandler := web.NewWebSocketHandler(globalContext)
	tokenHandler := web.NewTokenHandler(globalContext)

	router := httprouter.New()
	tokenHandler.InitWebHandler(router)
	webSocketHandler.InitWebHandler(router)
	router.GET("/ip", web.ReportIp)

	log.Infof("server will start on port %d", cfg.Server.Port)
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
		exception.Log("cmd/im/main.go: an exception caused on run server", err)
		return
	}
}
