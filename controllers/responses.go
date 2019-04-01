package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Error by APIdoc
type Error struct {
	Message string `json:"message"`
}

// SendError sends Error
func SendError(w http.ResponseWriter, msg string, status int) {
	SendResponse(w, Error{
		Message: msg,
	}, status)
}

// SendResponse sends response
func SendResponse(w http.ResponseWriter, v interface{}, status int) {
	log.Infof("status: %d, response: %v", status, v)
	w.Header().Set("Content-Type", "application/json")
	resp, err := json.Marshal(v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err = w.Write([]byte(fmt.Sprintf(`{"message": "%s"}`, err.Error()))); err != nil {
			log.Error(errors.Wrap(err, "unable to send error response"))
		}
		return
	}

	w.WriteHeader(status)
	if _, err = w.Write(resp); err != nil {
		log.Error(errors.Wrap(err, "unable to send response"))
	}
}
