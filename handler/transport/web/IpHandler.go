package web

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

func ReportIp(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if log.IsLevelEnabled(log.DebugLevel) {
		log.Debugln(r.Header)
	}

	ipAddress := r.Header.Get("X-Real-Ip")
	if ipAddress == "" {
		ipAddress = r.Header.Get("X-Forwarded-For")
	}
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}

	if _, err := w.Write([]byte(ipAddress)); err != nil {
		log.Errorf("failed to write remote addr: %s\n", err)
		return
	}
}
