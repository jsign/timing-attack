package measure

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"net/http/httptrace"
	"sort"
	"sync"
	"time"
)

// Result indicates results for a mesurement
type Result struct {
	Medians        []int64
	MaxMedianIndex int
	MaxMedian      int64
	BaseAvg        int64
	BaseStdDev     float64
}

var (
	transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 24 * time.Hour,
			DualStack: true,
		}).DialContext,
	}
)

// Measure latencies for reqs and return results
func Measure(reqs []http.Request, perReqCount int) (Result, error) {
	res := Result{}

	reqLatencies := make([][]int64, len(reqs))
	for i := range reqs {
		latencies, err := measureLatencies(&reqs[i], perReqCount)
		if err != nil {
			return res, fmt.Errorf("error while measuring latency in request %d: %w", i, err)
		}
		reqLatencies[i] = latencies
	}

	return calculateStats(reqLatencies), nil
}

func measureLatencies(r *http.Request, count int) ([]int64, error) {
	latencies := make([]int64, count)
	for i := 0; i < count; i++ {
		r := &http.Request{
			Method: r.Method,
			URL:    r.URL,
			Header: r.Header,
			Close:  true,
		}
		var wg sync.WaitGroup
		wg.Add(1)
		var latency time.Duration
		startTime := time.Now()
		tracer := &httptrace.ClientTrace{
			GotFirstResponseByte: func() {
				latency = time.Now().Sub(startTime)
				wg.Done()
			},
		}
		r = r.WithContext(httptrace.WithClientTrace(r.Context(), tracer))

		res, err := transport.RoundTrip(r)
		if err != nil {
			return nil, fmt.Errorf("error while doing roundtrip: %w", err)
		}
		err = res.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("error closing response body: %w", err)
		}

		wg.Wait()
		latencies[i] = int64(latency)
	}
	return latencies, nil
}

func calculateStats(reqLatencies [][]int64) Result {
	medians := calculateMedians(reqLatencies)

	var maxMedian int64
	var maxMedianIdx int
	for i, m := range medians {
		if m > maxMedian {
			maxMedian = m
			maxMedianIdx = i
		}
	}

	var sumLatenciesNonMax int64
	var count int64
	for i, latencies := range reqLatencies {
		if i == maxMedianIdx {
			continue
		}
		for _, v := range latencies {
			sumLatenciesNonMax += v
			count++
		}
	}
	baseAvgLatency := sumLatenciesNonMax / count

	var varianceSum int64
	for i, latencies := range reqLatencies {
		if i == maxMedianIdx {
			continue
		}
		for _, v := range latencies {
			varianceSum += (v - baseAvgLatency) * (v - baseAvgLatency)
		}
	}
	baseStdDeviation := math.Sqrt(float64(varianceSum / count))

	return Result{
		Medians:        medians,
		MaxMedianIndex: maxMedianIdx,
		MaxMedian:      maxMedian,
		BaseAvg:        baseAvgLatency,
		BaseStdDev:     baseStdDeviation,
	}
}

func calculateMedians(reqLatencies [][]int64) []int64 {
	res := make([]int64, len(reqLatencies))
	for i := range reqLatencies {
		count := len(reqLatencies[i])
		sort.Slice(reqLatencies[i], func(k, j int) bool { return reqLatencies[i][k] < reqLatencies[i][j] })
		if count&1 == 0 {
			res[i] = (reqLatencies[i][count/2-1] + reqLatencies[i][count/2]) / 2
			continue
		}
		res[i] = reqLatencies[i][count/2]
	}
	return res
}
