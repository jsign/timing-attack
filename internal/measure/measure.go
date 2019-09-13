package measure

import (
	"context"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"

	"github.com/pkg/errors"
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
func Measure(req http.Request, count int, concurrency int) ([]int64, error) {
	reqLatencies := make([]int64, count)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workerRes := make(chan []int64, concurrency)
	errChan := make(chan error)
	defer close(workerRes)
	defer close(errChan)

	for i := 0; i < concurrency; i++ {
		go workerMeasure(ctx, req, workerRes, errChan, count/concurrency)
	}

	for i := 0; i < concurrency; i++ {
		select {
		case newData := <-workerRes:
			reqLatencies = append(reqLatencies, newData...)

		case err := <-errChan:
			cancel()
			return []int64{}, errors.Errorf("error while measuring latency in request %d: %w", i, err)
		}
	}

	return reqLatencies, nil
}

func workerMeasure(ctx context.Context, req http.Request, resChan chan<- []int64, errChan chan<- error, count int) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	latencies, err := measureLatencies(&req, count)
	if err != nil {
		errChan <- err
		return
	}

	resChan <- latencies
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
			return nil, errors.Errorf("error while doing roundtrip: %w", err)
		}
		io.Copy(ioutil.Discard, res.Body)
		err = res.Body.Close()
		if err != nil {
			return nil, errors.Errorf("error closing response body: %w", err)
		}

		wg.Wait()
		latencies[i] = int64(latency)
	}
	return latencies, nil
}
