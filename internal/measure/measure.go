package measure

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"
)

var (
	transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 24 * time.Hour,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:    100,
		IdleConnTimeout: 10 * time.Second,
	}
)

// Measure latencies for reqs and return results
func Measure(reqs []http.Request, perReqCount int, concurrency int) ([][]int64, error) {
	reqLatencies := make([][]int64, len(reqs))

	workerRes := make(chan [][]int64, concurrency)
	errChan := make(chan error)
	for i := 0; i < concurrency; i++ {
		go func() {
			data := make([][]int64, len(reqs))
			for i := range reqs {
				latencies, err := measureLatencies(&reqs[i], perReqCount/concurrency)
				if err != nil {
					errChan <- err
					return
				}
				data[i] = latencies
			}

			workerRes <- data
		}()
	}

	for i := 0; i < concurrency; i++ {
		select {
		case newData := <-workerRes:
			for i := range newData {
				reqLatencies[i] = append(reqLatencies[i], newData[i]...)
			}
		case err := <-errChan:
			return [][]int64{}, fmt.Errorf("error while measuring latency in request %d: %w", i, err)
		}
	}

	return reqLatencies, nil
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
		io.Copy(ioutil.Discard, res.Body)
		err = res.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("error closing response body: %w", err)
		}

		wg.Wait()
		latencies[i] = int64(latency)
	}
	return latencies, nil
}
