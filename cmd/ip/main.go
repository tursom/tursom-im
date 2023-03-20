package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/ip", func(writer http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Header)

		IPAddress := r.Header.Get("X-Real-Ip")
		if IPAddress == "" {
			IPAddress = r.Header.Get("X-Forwarded-For")
		}
		if IPAddress == "" {
			IPAddress = r.RemoteAddr
		}

		if _, err := writer.Write([]byte(IPAddress)); err != nil {
			fmt.Errorf("failed to write remote addr: %s", err)
			return
		}
	})
	if err := http.ListenAndServe(fmt.Sprintf(":%s", os.Args[1]), nil); err != nil {
		panic(err)
	}
}
