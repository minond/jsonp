package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type JsonpRequeset struct {
	url         string
	method      string
	contentType string
	body        string
	callback    string
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", proxy)
	http.ListenAndServe(":"+port, nil)
}

func jsonpReq(r *http.Request) JsonpRequeset {
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

	return JsonpRequeset{url, method, contentType, body, callback}
}

func proxy(w http.ResponseWriter, r *http.Request) {
	jsonp := jsonpReq(r)

	if jsonp.url == "" {
		http.Error(w, "Missing a `url` query parameter.", http.StatusBadRequest)
		return
	}

	var res *http.Response
	var err error

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

	if err != nil {
		http.Error(w, "Error making proxy request.", http.StatusInternalServerError)
		return
	}

	json := buffRead(res.Body)
	body := fmt.Sprintf("%s(%s)", jsonp.callback, json)

	w.Header().Set("Content-Type", "application/javascript")
	w.WriteHeader(res.StatusCode)
	w.Write([]byte(body))
}

func buffRead(r io.ReadCloser) string {
	defer r.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)

	return buf.String()
}
