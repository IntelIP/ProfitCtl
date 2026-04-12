package simulation

import (
	"testing"

	"github.com/IntelIP/ProfitCtl/pkg/types"
)

func BenchmarkRunMonteCarlo(b *testing.B) {
	meanCompute := 10.0
	stdDevCompute := 1.0
	meanStorage := 5.0
	stdDevStorage := 0.25

	fixedCosts := []types.FixedCost{
		{Name: "Infrastructure", Amount: 700.0, Period: types.PeriodMonthly, Layer: types.LayerInfrastructure},
		{Name: "Application", Amount: 300.0, Period: types.PeriodMonthly, Layer: types.LayerApplication},
		{Name: "Service", Amount: 100.0, Period: types.PeriodMonthly, Layer: types.LayerService},
	}

	variableCosts := []types.VariableCost{
		{Name: "Compute", CostPerUnit: 10.0, UnitsPerUser: 1.0, Distribution: types.DistNormal, Mean: &meanCompute, StdDev: &stdDevCompute, Layer: types.LayerInfrastructure},
		{Name: "Storage", CostPerUnit: 5.0, UnitsPerUser: 1.0, Distribution: types.DistNormal, Mean: &meanStorage, StdDev: &stdDevStorage, Layer: types.LayerInfrastructure},
	}

	users := 100
	months := 12

	config := NewMonteCarloConfig(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RunMonteCarlo(fixedCosts, variableCosts, users, months, config)
	}
}

func BenchmarkRunStressTest(b *testing.B) {
	meanCompute := 10.0
	stdDevCompute := 1.0
	meanStorage := 5.0
	stdDevStorage := 0.25

	fixedCosts := []types.FixedCost{
		{Name: "Infrastructure", Amount: 700.0, Period: types.PeriodMonthly, Layer: types.LayerInfrastructure},
		{Name: "Application", Amount: 300.0, Period: types.PeriodMonthly, Layer: types.LayerApplication},
		{Name: "Service", Amount: 100.0, Period: types.PeriodMonthly, Layer: types.LayerService},
	}

	variableCosts := []types.VariableCost{
		{Name: "Compute", CostPerUnit: 10.0, UnitsPerUser: 1.0, Distribution: types.DistNormal, Mean: &meanCompute, StdDev: &stdDevCompute, Layer: types.LayerInfrastructure},
		{Name: "Storage", CostPerUnit: 5.0, UnitsPerUser: 1.0, Distribution: types.DistNormal, Mean: &meanStorage, StdDev: &stdDevStorage, Layer: types.LayerInfrastructure},
	}

	users := 100
	months := 12
	iterations := 1000

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RunStressTest(fixedCosts, variableCosts, users, months, iterations)
	}
	b.Logf("Ran %d stress tests", b.N)
}

func BenchmarkRunMonteCarloLarge(b *testing.B) {
	meanCompute := 100.0
	stdDevCompute := 20.0
	meanStorage := 50.0
	stdDevStorage := 5.0
	meanBandwidth := 25.0
	stdDevBandwidth := 3.75

	fixedCosts := []types.FixedCost{
		{Name: "Infrastructure", Amount: 7000.0, Period: types.PeriodMonthly, Layer: types.LayerInfrastructure},
		{Name: "Application", Amount: 3000.0, Period: types.PeriodMonthly, Layer: types.LayerApplication},
		{Name: "Service", Amount: 1000.0, Period: types.PeriodMonthly, Layer: types.LayerService},
	}

	variableCosts := []types.VariableCost{
		{Name: "Compute", CostPerUnit: 100.0, UnitsPerUser: 1.0, Distribution: types.DistNormal, Mean: &meanCompute, StdDev: &stdDevCompute, Layer: types.LayerInfrastructure},
		{Name: "Storage", CostPerUnit: 50.0, UnitsPerUser: 1.0, Distribution: types.DistNormal, Mean: &meanStorage, StdDev: &stdDevStorage, Layer: types.LayerInfrastructure},
		{Name: "Bandwidth", CostPerUnit: 25.0, UnitsPerUser: 1.0, Distribution: types.DistNormal, Mean: &meanBandwidth, StdDev: &stdDevBandwidth, Layer: types.LayerApplication},
	}

	users := 10000
	months := 24

	config := NewMonteCarloConfig(10000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RunMonteCarlo(fixedCosts, variableCosts, users, months, config)
	}
}
