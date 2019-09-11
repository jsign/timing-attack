package main

import (
	"flag"
	"net/http"
	"net/url"
	"os"

	"github.com/jsign/timing-attack/internal/measure"
	"github.com/jsign/timing-attack/internal/stats"
	log "github.com/sirupsen/logrus"
)

const (
	minIterations = 10
)

var (
	addr        = flag.String("serveraddr", "http://localhost:3001", "server address")
	debug       = flag.Bool("debug", false, "debug mode")
	maxIterScan = flag.Int("maxIterScan", 10000, "maximum iterations during scan")
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

	reqs := generateRequests(cases)
	accumulatedData := make([][]int64, len(cases))
	accumulatedIterations := 0
	for it := minIterations; it <= *maxIterScan; it += minIterations {
		newData, err := measure.Measure(reqs, it)
		if err != nil {
			logger.Fatalf("error while measuring test cases: %+v", err)
		}
		for i := range newData {
			accumulatedData[i] = append(accumulatedData[i], newData[i]...)
		}
		accumulatedIterations += it

		logger.Debugf("Total of %v iterations:\n", accumulatedIterations)
		printStats(logger, cases, stats.Calculate(accumulatedData))

	}
}

func generateRequests(cases []string) []http.Request {
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
	return reqs
}

func printStats(logger *log.Logger, cases []string, s stats.Stats) {
	logger.Debugf("\tMax median latency: %s in %.2fms", cases[s.MaxMedianIndex], float64(s.MaxMedian)/1000000)
	logger.Debugf("\tBase average latency is: %.2fms", float64(s.BaseAvg)/1000000)
	logger.Debugf("\tBase stddev is: %.2fms", float64(s.BaseStdDev)/1000000)
	for i := range s.Medians {
		latencyRatio := float64(s.Medians[i]-s.BaseAvg) / float64(s.BaseStdDev) * 100
		logger.Debugf("\tMedian latency for %s is %.2fms (%.2f%%)", cases[i], float64(s.Medians[i])/1000000, latencyRatio)
	}
}
