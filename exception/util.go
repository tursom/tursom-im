package exception

import (
	log "github.com/sirupsen/logrus"
	"github.com/tursom/GoCollections/exceptions"
)

func Log(msg string, err error) {
	if err == nil {
		return
	}

	switch err.(type) {
	case exceptions.Exception:
		log.WithField("stack", exceptions.GetStackTraceString(err.(exceptions.Exception))).
			Error(msg)
	default:
		log.WithField("err", err).
			Error(msg)
	}
}
