package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/skratchdot/open-golang/open"

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

	if err := open.Run(fmt.Sprintf("http://localhost:%d", *port)); err != nil {
		log.Err(err).Msg("could not open browser")
	}

	<-ch
}
