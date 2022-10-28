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
	"github.com/tursom-im/context"
	"github.com/tursom-im/handler"
	"github.com/tursom-im/handler/msg"
)

func systemInit() *config.Config {
	msg.Init()
	rand.Seed(time.Now().UnixNano())

	cfg := config.NewConfig()
	configFile, err := os.ReadFile("config.yaml")
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
	//node, err := libp2p.New(libp2p.ChainOptions())
	//if err != nil {
	//	panic(err)
	//}
	//// 打印节点的所有地址
	//fmt.Println("Listen addresses:", node.Addrs())
	//// 关闭节点，然后退出
	//if err = node.Close(); err != nil {
	//	panic(err)
	//}
	//return

	cfg := systemInit()
	fmt.Println(cfg)

	globalContext := context.NewGlobalContext(cfg)
	if globalContext == nil {
		panic(exceptions.NewRuntimeException("", nil))
	}

	webSocketHandler := handler.NewWebSocketHandler(globalContext)
	tokenHandler := handler.NewTokenHandler(globalContext)

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
