package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	w.Header().Set("Content-Type", "application/json")
	resp, err := json.Marshal(v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"message": "%s"}`, err.Error())))
		return
	}
	w.WriteHeader(status)
	fmt.Println(resp)
	w.Write(resp)
}
