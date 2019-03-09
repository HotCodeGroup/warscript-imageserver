package main

import (
	"net/http"
	"os"

	"github.com/HotCodeGroup/warscript-imageserver/controllers"
	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
}

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/photos", controllers.UploadPhoto).Methods("POST")
	r.HandleFunc("/photos/{photo_uuid}", controllers.GetPhoto).Methods("GET")
	r.HandleFunc("/", controllers.MainPage).Methods("GET")
	port := os.Getenv("PORT")
	log.Infof("MainService successfully started at port %s", port)
	err := http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Errorf("cant start main server. err: %s", err.Error())
		return
	}
}
