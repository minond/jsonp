package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type JsonpRequest struct {
	url         string
	method      string
	contentType string
	body        string
	callback    string
}

type HttpResponse struct {
	res *http.Response
	err error
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", proxy)
	http.HandleFunc("/help", help)
	http.ListenAndServe(":"+port, nil)
}

func jsonpReq(r *http.Request) JsonpRequest {
	url := r.URL.Query().Get("url")
	method := r.URL.Query().Get("method")
	contentType := r.URL.Query().Get("contentType")
	body := r.URL.Query().Get("body")
	callback := r.URL.Query().Get("callback")

	if contentType == "" {
		contentType = "application/json"
	}

	if callback == "" {
		callback = "callback"
	}

	if method == "" {
		method = http.MethodGet
	}

	return JsonpRequest{url, method, contentType, body, callback}
}

func buffRead(r io.ReadCloser) string {
	defer r.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)

	return buf.String()
}

func proxy(w http.ResponseWriter, r *http.Request) {
	jsonp := jsonpReq(r)

	if jsonp.url == "" {
		http.Error(w, "Missing a `url` query parameter.", http.StatusBadRequest)
		return
	}

	ch := make(chan *HttpResponse)
	start := time.Now()

	go func() {
		var res *http.Response
		var err error

		fmt.Printf("Sending a %s request to %s\n", jsonp.method, jsonp.url)

		switch strings.ToUpper(jsonp.method) {
		case http.MethodPost:
			res, err = http.Post(
				jsonp.url,
				jsonp.contentType,
				strings.NewReader(jsonp.body),
			)

		default:
			res, err = http.Get(jsonp.url)
		}

		ch <- &HttpResponse{res, err}
	}()

	for {
		select {
		case r := <-ch:
			secs := time.Since(start).Seconds()
			fmt.Printf("Got a response back from %s in %.5f seconds\n", jsonp.url, secs)

			if r.err != nil {
				http.Error(w, "Error making proxy request.", http.StatusInternalServerError)
				return
			}

			json := buffRead(r.res.Body)
			body := fmt.Sprintf("%s(%s)", jsonp.callback, json)

			w.Header().Set("Content-Type", "application/javascript")
			w.WriteHeader(r.res.StatusCode)
			w.Write([]byte(body))

			return

		case <-time.After(200 * time.Millisecond):
			secs := time.Since(start).Seconds()
			fmt.Printf("Still waiting for %s after %.5f seconds\n", jsonp.url, secs)
		}
	}
}

func help(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(`Parameters
		- url (string) the url you need to request
		- method (string) http method to use when making request
		- contentType (string) type of payload (if 'body' is set)
		- body (string) payload to send as request body (if 'method' is POST)
		- callback (string) name of callback function
	`))
}
