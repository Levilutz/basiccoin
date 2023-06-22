package rest

import (
	"io"
	"net/http"
)

// func write400(w http.ResponseWriter, r *http.Request, err error) {
// 	w.WriteHeader(400)
// 	io.WriteString(w, "client error: "+err.Error())
// }

func write401(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(401)
	io.WriteString(w, "unauthorized")
}

// func write422(w http.ResponseWriter, r *http.Request, err error) {
// 	w.WriteHeader(422)
// 	io.WriteString(w, "failed to parse: "+err.Error())
// }

func write405(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(405)
	io.WriteString(w, "method not allowed: "+r.Method)
}

// func write500(w http.ResponseWriter, r *http.Request, err error) {
// 	w.WriteHeader(500)
// 	io.WriteString(w, "server error: "+err.Error())
// }
