package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/gosuri/uilive"
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
	concurreny  = flag.Int("concurrency", 8, "concurrency level to make measurements")
)

func main() {
	flag.Parse()

	logger := log.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(log.InfoLevel)
	if *debug {
		logger.SetLevel(log.DebugLevel)
	}
	writer := uilive.New()
	writer.Start()
	defer writer.Stop()

	baseCase := "whatever@fake.com"
	targetCase := "correct@email.com"

	reqs := generateRequests([]string{baseCase, targetCase})
	baseReq, targetReq := reqs[0], reqs[1]

	var baseData, targetData []int64
	accumulatedIterations := 0
	for it := minIterations; it <= *maxIterScan; it += minIterations {
		newBaseData, err := measure.Measure(baseReq, it, *concurreny)
		newTargetData, err := measure.Measure(targetReq, it, *concurreny)
		if err != nil {
			logger.Fatalf("error while measuring test cases: %+v", err)
		}
		baseData = append(baseData, newBaseData...)
		targetData = append(targetData, newTargetData...)
		accumulatedIterations += it

		s, err := stats.Calculate(baseData, targetData)
		if err != nil {
			logger.Fatalf("error while calculating stats: %v", err)
		}
		printStats(writer, accumulatedIterations, s)
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

func printStats(w *uilive.Writer, iterations int, s stats.Stats) {
	fmt.Fprintf(w, "%4d iterations | Base Mean Latency CI (%.3f, %.3f) | Target Mean Latency CI: (%.3f, %.3f) | CI at 95%%\n", iterations, s.BaseCI.Left, s.BaseCI.Right, s.TargetCI.Left, s.TargetCI.Right)
}
