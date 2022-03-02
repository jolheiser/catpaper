package main

import (
	"flag"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	port := flag.Int("port", 8080, "The port to serve on")
	flag.Parse()

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	cache := filepath.Join(cacheDir, "catpaper")

	if err := download(cache); err != nil {
		log.Err(err).Msg("")
	}

	go server(cache, *port)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Kill, os.Interrupt)
	<-ch
}
