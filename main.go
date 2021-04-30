package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
	"tursom-im/config"
	"tursom-im/context"
	"tursom-im/handler"
)

func main() {
	cfg := config.NewConfig()
	configFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return
	}
	fmt.Println(string(configFile))
	err = yaml.Unmarshal(configFile, cfg)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(cfg)

	rand.Seed(time.Now().UnixNano())

	globalContext := context.NewGlobalContext(cfg)
	webSocketHandler := handler.NewWebSocketHandler(globalContext)

	router := httprouter.New()
	router.GET("/ws", webSocketHandler.UpgradeToWebSocket)
	http.ListenAndServe(":12345", router)
}
