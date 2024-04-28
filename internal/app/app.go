package app

import (
	"log"
	"net/http"
)

func Mux(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		log.Println("GET method requested", r.RemoteAddr)
	case http.MethodPost:
		log.Println("POST method requested", r.RemoteAddr)
	default:
		log.Println("Unknown method requested", r.RemoteAddr)
	}
}
