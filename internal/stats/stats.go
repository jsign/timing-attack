package stats

import (
	"math"
	"sort"
)

// Stats indicates statistics for latency data
type Stats struct {
	Medians        []int64
	MaxMedianIndex int
	MaxMedian      int64
	BaseAvg        int64
	BaseStdDev     float64
}

// Calculate calculate stats for latency data
func Calculate(reqLatencies [][]int64) Stats {
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

	return Stats{
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
