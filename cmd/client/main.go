package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func main() {
	endpoint := "http://localhost:8080/"
	// Data container for the request
	data := url.Values{}
	// Console prompt
	fmt.Println("Enter the long URL:")
	// Open streaming read from the console
	reader := bufio.NewReader(os.Stdin)
	// Read a line from the console
	long, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	long = strings.TrimSuffix(long, "\n")
	// Fill the data container with the input
	data.Set("url", long)
	// Constructing the HTTP client
	client := &http.Client{}
	// Constructing the request
	// A POST request should, besides headers, contain a body
	// The body should be a stream reader io.Reader
	// In most cases, bytes.Buffer works perfectly
	request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBufferString(data.Encode()))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// In the request headers, specify that the data is encoded with standard URL scheme
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	// Send the request and receive the response
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Print the response code
	fmt.Println("Status Code:", response.Status)
	defer response.Body.Close()
	// Read the stream from the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Print the response body
	fmt.Println(string(body))
}
