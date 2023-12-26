package routes

import (
	"log"
	"net/http"

	"github.com/Charith01/CustomerLabs/app"
	"github.com/gorilla/mux"
)

func InitializeRoutes() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/sendEvent", app.SendEventToWorker).Methods("POST")
	log.Fatal(http.ListenAndServe(":8000", router))

}
