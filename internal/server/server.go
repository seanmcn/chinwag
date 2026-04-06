// Package server serves the rendered report on a local http port.
package server

import (
	"fmt"
	"net/http"

	"github.com/seanmcn/whatsapp-analyse/internal/analyse"
	"github.com/seanmcn/whatsapp-analyse/internal/render"
)

func Serve(addr string, s analyse.Stats) error {
	html, err := render.HTML(s)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(html)
	})
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(render.StaticFS()))))
	fmt.Printf("Serving on http://%s\n", addr)
	return http.ListenAndServe(addr, mux)
}
