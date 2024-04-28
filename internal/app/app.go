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
	log.Printf("%s method requested by %s", r.Method, r.RemoteAddr)
	switch r.Method {
	case http.MethodGet:
		a.expand(w, r)
	case http.MethodPost:
		a.shorten(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (a *App) shorten(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	url, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatalln("error")
	}

	hash := createHash(string(url))
	w.WriteHeader(http.StatusCreated)
	w.Header().Add("Content-Type", "text/plain")
	w.Write([]byte(hash))
}

func (a *App) expand(w http.ResponseWriter, r *http.Request) {
}

func createHash(url string) string {
	log.Printf("original url %s", url)
	return "shortUrl"
}
