package base

import (
	"net/http"

	"github.com/a-h/templ"
)

func Serve(component templ.Component, w http.ResponseWriter) http.HandlerFunc {
	w.Header().Set("HX-Trigger", "successful-event")
	return templ.Handler(component, templ.WithStatus(http.StatusOK)).ServeHTTP
}

func ServeNoSwap(w http.ResponseWriter) http.HandlerFunc {
	w.Header().Set("HX-Trigger", "successful-event")
	return func(w http.ResponseWriter, r *http.Request) {}
}

func ErrorServe(error string, status int, w http.ResponseWriter) http.HandlerFunc {
	w.Header().Set("HX-Trigger", `{"unsuccessful-event": "`+error+`" }`)
	w.WriteHeader(status)
	return func(w http.ResponseWriter, r *http.Request) {}
}

func ErrorServeRedirect(error string, status int, w http.ResponseWriter) http.HandlerFunc {
	w.Header().Set("HX-Redirect", "/projects")
	w.Header().Set("HX-Trigger", `{"unsuccessful-event": "`+error+`" }`)
	w.Header().Set("HX-Reswap", "none")
	w.WriteHeader(status)
	return func(w http.ResponseWriter, r *http.Request) {}
}
