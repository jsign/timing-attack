package stats

import (
	"math"

	"github.com/montanaflynn/stats"
	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
)

const (
	c = float64(0.95)
)

var (
	normalDistribution = distuv.Normal{
		Mu:    0,
		Sigma: 1,
		Src:   rand.NewSource(uint64(0)),
	}
)

// Stats indicates statistics for latency data
type Stats struct {
	BaseCI   ConfidenceInterval
	TargetCI ConfidenceInterval
}

// ConfidenceInterval is a confidence interval for a sample data
type ConfidenceInterval struct {
	Left  float64
	Right float64
	C     float64
}

// Calculate calculate stats for latency data
func Calculate(base, target []int64) (res Stats, err error) {
	baseCI, err := calculateCI(toFloat64(base))
	if err != nil {
		return
	}
	targetCI, err := calculateCI(toFloat64(target))
	if err != nil {
		return
	}

	res.BaseCI = baseCI
	res.TargetCI = targetCI
	return
}

func calculateCI(data []float64) (res ConfidenceInterval, err error) {
	sampleMean, err := stats.Mean(stats.Float64Data(data))
	if err != nil {
		return
	}
	stdev, err := stats.StandardDeviationSample(stats.Float64Data(data))
	if err != nil {
		return
	}
	t := normalDistribution.Quantile((1 - c) / 2)

	delta := t * stdev / math.Sqrt(float64(len(data)))
	res.Left = sampleMean + delta
	res.Right = sampleMean - delta
	res.C = c

	return
}

func toFloat64(data []int64) []float64 {
	res := make([]float64, len(data))
	for i := range data {
		res[i] = float64(data[i] / 1000000)
	}
	return res
}
