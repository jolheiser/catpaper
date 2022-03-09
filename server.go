package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

var (
	//go:embed assets/index.tmpl
	_tmpl string
	tmpl  = template.Must(template.New("").Parse(_tmpl))
	//go:embed assets/favicon.ico
	faviconIco []byte
)

func server(dir string, port int) {
	http.HandleFunc("/favicon.ico", favicon)
	http.Handle("/wallpaper/", http.StripPrefix("/wallpaper", http.FileServer(http.Dir(dir))))

	wp, err := wallpapers(dir)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	http.HandleFunc("/", index(wp))

	log.Info().Msgf("listening at http://localhost:%d", port)
	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func index(wallpapers map[string][]wallpaper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := tmpl.Execute(w, wallpapers); err != nil {
			log.Err(err).Msg("")
		}
	}
}

func favicon(w http.ResponseWriter, r *http.Request) {
	http.ServeContent(w, r, "favicon.ico", time.Time{}, bytes.NewReader(faviconIco))
}
