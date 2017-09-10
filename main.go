package main

import (
	"bytes"
	"io"
	"net/http"
	"strings"
)

type JsonpRequeset struct {
	url         string
	method      string
	contentType string
	body        string
}

func main() {
	http.HandleFunc("/", proxy)
	http.ListenAndServe(":8080", nil)
}

func jsonpReq(r *http.Request) JsonpRequeset {
	url := r.URL.Query().Get("url")
	method := r.URL.Query().Get("method")
	contentType := r.URL.Query().Get("contentType")
	body := r.URL.Query().Get("body")

	if contentType == "" {
		contentType = "application/json"
	}

	if method == "" {
		method = http.MethodGet
	}

	return JsonpRequeset{url, method, contentType, body}
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

	w.WriteHeader(res.StatusCode)
	w.Write(buffBytes(res.Body))
}

func buffBytes(r io.ReadCloser) []byte {
	defer r.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)

	return buf.Bytes()
}
