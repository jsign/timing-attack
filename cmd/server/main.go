package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jsign/timing-attack/cmd/server/handler"
	log "github.com/sirupsen/logrus"
)

var (
	port        = flag.Int("port", 3001, "port to listen to")
	stdDev      = flag.Int("stddev", 0, "noise standard deviation")
	baseLatency = flag.Int("baseLatency", 0, "noise base latency")
	debug       = flag.Bool("debug", false, "debug mode")
)

func main() {
	flag.Parse()

	logger := log.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(log.InfoLevel)
	if *debug {
		logger.SetLevel(log.DebugLevel)
	}

	logger.Debugf("Base latency: %dms", *baseLatency)
	logger.Debugf("Latency stdev: %dms", *stdDev)

	s := http.Server{
		Addr:    fmt.Sprintf("localhost:%d", *port),
		Handler: handler.NewNaiveComparator(logger, *baseLatency, *stdDev),
	}
	go func() {
		logger.Debug("webserver listening...")
		if err := s.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatalf("error while listening: %v", err)
			}
		}
	}()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	if err := s.Shutdown(context.Background()); err != nil {
		logger.Fatalf("coudn't shutdown the server: %v", err)
	}
	logger.Debug("server shutdown successfully")
}
