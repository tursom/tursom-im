package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/tursom/tursom-im/handler/transport/web"
)

func main() {
	http.HandleFunc("/ip", func(writer http.ResponseWriter, r *http.Request) {
		web.ReportIp(writer, r, nil)
	})
	if err := http.ListenAndServe(fmt.Sprintf(":%s", os.Args[1]), nil); err != nil {
		panic(err)
	}
}
