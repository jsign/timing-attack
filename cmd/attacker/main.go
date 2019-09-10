package main

import (
	"flag"
	"net/http"
	"net/url"
	"os"

	"github.com/jsign/timing-attack/internal/measure"
	log "github.com/sirupsen/logrus"
)

var (
	addr       = flag.String("serveraddr", "http://localhost:3001", "server address")
	debug      = flag.Bool("debug", false, "debug mode")
	iterations = flag.Int("iter", 100, "iterations to execute per case")
)

func main() {
	flag.Parse()

	logger := log.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(log.InfoLevel)
	if *debug {
		logger.SetLevel(log.DebugLevel)
	}

	cases := []string{
		"correct@email.com",
		"whatever@fake.com",
		"foo@fake.com",
	}

	reqs := make([]http.Request, len(cases))
	for i := range cases {
		serverURL, err := url.Parse(*addr)
		if err != nil {
			log.Fatalf("server url is invalid: %v", err)
		}
		serverURL.RawQuery = url.Values{"email": []string{cases[i]}}.Encode()
		reqs[i] = http.Request{
			Method: http.MethodGet,
			URL:    serverURL,
			Header: http.Header{},
		}
	}

	res, err := measure.Measure(reqs, *iterations)
	if err != nil {
		logger.Fatalf("error while measuring test cases: %+v", err)
	}

	logger.Debugf("Max median latency: %s in %.2fms", cases[res.MaxMedianIndex], float64(res.MaxMedian)/1000000)
	logger.Debugf("Base average latency is: %.2fms", float64(res.BaseAvg)/1000000)
	logger.Debugf("Base stddev is: %.2fms", float64(res.BaseStdDev)/1000000)
	for i := range res.Medians {
		latencyRatio := float64(res.Medians[i]-res.BaseAvg) / float64(res.BaseStdDev) * 100
		logger.Debugf("Median latency for %s is %.2fms (%.2f%%)", cases[i], float64(res.Medians[i])/1000000, latencyRatio)
	}
}
