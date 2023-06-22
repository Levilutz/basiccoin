package rest

import (
	"fmt"
	"io"
	"net/http"
)

// Get a single query string value, return error if too many / too few.
func getSingleQueryString(w http.ResponseWriter, r *http.Request, name string) (string, error) {
	values, ok := r.URL.Query()[name]
	if !ok || len(values) == 0 {
		return "", fmt.Errorf("expect 1 value of '%s', received 0", name)
	} else if len(values) > 1 {
		return "", fmt.Errorf("expect 1 value of '%s', received %d", name, len(values))
	}
	return values[0], nil
}

func write400(w http.ResponseWriter, err error) {
	w.WriteHeader(400)
	io.WriteString(w, "client error: "+err.Error())
}

func write401(w http.ResponseWriter) {
	w.WriteHeader(401)
	io.WriteString(w, "unauthorized")
}

// func write422(w http.ResponseWriter) {
// 	w.WriteHeader(422)
// 	io.WriteString(w, "failed to parse: "+err.Error())
// }

func write405(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(405)
	io.WriteString(w, "method not allowed: "+r.Method)
}

func write500(w http.ResponseWriter, err error) {
	w.WriteHeader(500)
	io.WriteString(w, "server error: "+err.Error())
}
