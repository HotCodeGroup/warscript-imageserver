package main

import (
	"net/http"
	"os"

	"github.com/HotCodeGroup/warscript-imageserver/controllers"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
}

func main() {
	control, err := controllers.Init(os.Getenv("AWS_ACCESS_KEY_ID"),
		os.Getenv("AWS_SECRET_ACCESS_KEY"), "", "images.warscript")
	if err != nil {
		log.Errorf("cant start main server: storage can't be loaded: err: %s", err.Error())
		return
	}

	r := mux.NewRouter()
	r.HandleFunc("/photos", control.UploadPhoto).Methods("POST")
	r.HandleFunc("/photos/{photo_uuid}", control.GetPhoto).Methods("GET")
	r.HandleFunc("/", controllers.MainPage).Methods("GET")
	corsMiddleware := handlers.CORS(
		handlers.AllowedOrigins([]string{os.Getenv("CORS_HOST")}),
		handlers.AllowedMethods([]string{"POST", "GET", "PUT", "DELETE"}),
		handlers.AllowedHeaders([]string{"Content-Type"}),
		handlers.AllowCredentials(),
	)

	port := os.Getenv("PORT")
	log.Infof("MainService successfully started at port %s", port)
	err = http.ListenAndServe(":"+port, corsMiddleware(r))
	if err != nil {
		log.Errorf("cant start main server: err: %s", err.Error())
		return
	}
}
