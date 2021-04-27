package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"tursom-im/context"
	"tursom-im/handler"
)

func main() {
	globalContext := context.GlobalContext{}
	webSocketHandler := handler.NewWebSocketHandler(globalContext)

	router := httprouter.New()
	router.GET("/ws", webSocketHandler.UpgradeToWebSocket)
	http.ListenAndServe(":12345", router)
}
