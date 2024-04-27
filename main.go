package main

import "net/http"

func HelloWorld(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<h1>Hello, World</h1>"))
}

func main() {
	http.HandleFunc("/", HelloWorld)
	http.ListenAndServe(":8080", nil)
}
