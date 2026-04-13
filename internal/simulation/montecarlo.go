package simulation

import (
	"math"
	"sort"
	"time"

	"github.com/IntelIP/ProfitCtl/internal/config"
	"github.com/IntelIP/ProfitCtl/internal/cost"
	"github.com/IntelIP/ProfitCtl/internal/pricing"
	"github.com/IntelIP/ProfitCtl/pkg/types"
)

type MonteCarloResult struct {
	Iterations int
	Samples    []float64
	P95        float64
	P99        float64
	Mean       float64
	Median     float64
	StdDev     float64
}

type MonteCarloConfig struct {
	Iterations int
	Seed       int64
}

func NewMonteCarloConfig(iterations int) MonteCarloConfig {
	if iterations <= 0 {
		iterations = 10000
	}
	return MonteCarloConfig{
		Iterations: iterations,
		Seed:       time.Now().UnixNano(),
	}
}

func RunMonteCarlo(fixedCosts []types.FixedCost, variableCosts []types.VariableCost, pricingCfg *config.PricingConfig, users int, months int, config MonteCarloConfig) MonteCarloResult {
	samples := make([]float64, 0, config.Iterations)

	fixedMonthly := cost.CalculateFixedCosts(fixedCosts, 1)
	fixedTotal := fixedMonthly.Total * float64(months)
	paidUsers := users
	if pricingCfg != nil {
		paidUsers = pricing.CalculateRevenue(pricingCfg, users).PaidUsers
	}

	for i := 0; i < config.Iterations; i++ {
		variableResult := cost.CalculateVariableCostWithVariabilityAndScope(variableCosts, users, paidUsers)
		totalCost := fixedTotal + variableResult.Total
		samples = append(samples, totalCost)
	}

	sort.Float64s(samples)

	result := MonteCarloResult{
		Iterations: config.Iterations,
		Samples:    samples,
		P95:        calculatePercentile(samples, 0.95),
		P99:        calculatePercentile(samples, 0.99),
		Mean:       calculateMean(samples),
		Median:     calculateMedian(samples),
		StdDev:     calculateStdDev(samples),
	}

	return result
}

func calculatePercentile(sortedSamples []float64, percentile float64) float64 {
	if len(sortedSamples) == 0 {
		return 0
	}

	index := int(percentile * float64(len(sortedSamples)))
	if index >= len(sortedSamples) {
		index = len(sortedSamples) - 1
	}

	return sortedSamples[index]
}

func calculateMean(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}

	sum := 0.0
	for _, sample := range samples {
		sum += sample
	}

	return sum / float64(len(samples))
}

func calculateMedian(sortedSamples []float64) float64 {
	if len(sortedSamples) == 0 {
		return 0
	}

	n := len(sortedSamples)
	if n%2 == 0 {
		return (sortedSamples[n/2-1] + sortedSamples[n/2]) / 2
	}

	return sortedSamples[n/2]
}

func calculateStdDev(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}

	mean := calculateMean(samples)
	sumSquaredDiff := 0.0

	for _, sample := range samples {
		diff := sample - mean
		sumSquaredDiff += diff * diff
	}

	variance := sumSquaredDiff / float64(len(samples))
	return math.Sqrt(variance)
}

type StressTestResult struct {
	Users              int
	Months             int
	MonteCarlo         MonteCarloResult
	MeanCostPerUser    float64
	P95CostPerUser     float64
	P99CostPerUser     float64
	WorstCaseTotalCost float64
}

func RunStressTest(fixedCosts []types.FixedCost, variableCosts []types.VariableCost, pricingCfg *config.PricingConfig, users int, months int, iterations int) StressTestResult {
	mcConfig := NewMonteCarloConfig(iterations)
	monteCarlo := RunMonteCarlo(fixedCosts, variableCosts, pricingCfg, users, months, mcConfig)

	meanCostPerUser := monteCarlo.Mean / float64(users)
	p95CostPerUser := monteCarlo.P95 / float64(users)
	p99CostPerUser := monteCarlo.P99 / float64(users)

	return StressTestResult{
		Users:              users,
		Months:             months,
		MonteCarlo:         monteCarlo,
		MeanCostPerUser:    roundToPrecision(meanCostPerUser, 2),
		P95CostPerUser:     roundToPrecision(p95CostPerUser, 2),
		P99CostPerUser:     roundToPrecision(p99CostPerUser, 2),
		WorstCaseTotalCost: roundToPrecision(monteCarlo.P99, 2),
	}
}

func roundToPrecision(value float64, precision int) float64 {
	multiplier := math.Pow(10, float64(precision))
	return math.Round(value*multiplier) / multiplier
}
