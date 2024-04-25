package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"
)

func main() {
	client := &http.Client{}
	client.Timeout = time.Second * 1
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 2 {
			return errors.New("Stopped after 2 redirects")
		}
		return nil
	}

	request, err := http.NewRequest(http.MethodGet, "https://research.swtch.com/interfaces", nil)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Add("Accept", "application/json")
	if err != nil {
		fmt.Println(err)
	}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
	}

	contentType := response.Header.Get("content-type")

	defer response.Body.Close()

	requestDump, err := httputil.DumpRequest(request, true)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(requestDump))

	fmt.Println(response.Status)
	fmt.Println(contentType)

}
