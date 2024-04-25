package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	response, err := http.Get("http://example.com")
	if err != nil {
		fmt.Println(err)
	}
	contentType := response.Header.Get("content-type")

	defer response.Body.Close()
	payload, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(response.Status)
	fmt.Println(contentType)
	fmt.Println(string(payload))
}
