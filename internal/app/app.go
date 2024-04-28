package app

import (
	"io"
	"log"
	"net/http"
)

type App struct{}

func (a *App) Run() error {
	http.HandleFunc("/", a.Mux)
	return http.ListenAndServe(":8080", nil)
}

func (a *App) Mux(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.expand(w, r)
	case http.MethodPost:
		a.shorten(w, r)
	default:
		log.Println("Unknown method requested", r.RemoteAddr)
	}
}

func (a *App) shorten(w http.ResponseWriter, r *http.Request) {
	log.Println("GET method requested", r.RemoteAddr)

	defer r.Body.Close()
	url, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatalln("error")
	}

	hash := createHash(string(url))
	w.Write([]byte(hash))
}

func (a *App) expand(w http.ResponseWriter, r *http.Request) {
	log.Println("POST method requested", w, r.RemoteAddr)
}

func createHash(url string) string {
	log.Printf("original url %s", url)
	return "shortUrl"
}
