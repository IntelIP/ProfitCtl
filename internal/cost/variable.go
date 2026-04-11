package cost

import (
	"errors"
	"math"
	"math/rand"

	"github.com/IntelIP/ProfitCtl/pkg/types"
)

type DistributionGenerator func() float64

func NewDistributionGenerator(vc types.VariableCost) (DistributionGenerator, error) {
	switch vc.Distribution {
	case types.DistNormal:
		if vc.Mean == nil || vc.StdDev == nil {
			return nil, errors.New("normal distribution requires mean and stddev")
		}
		return func() float64 {
			return *vc.Mean + *vc.StdDev*rand.NormFloat64()
		}, nil

	case types.DistUniform:
		if vc.Min == nil || vc.Max == nil {
			return nil, errors.New("uniform distribution requires min and max")
		}
		return func() float64 {
			return *vc.Min + (*vc.Max-*vc.Min)*rand.Float64()
		}, nil

	case types.DistExponential:
		if vc.Rate == nil {
			return nil, errors.New("exponential distribution requires rate")
		}
		return func() float64 {
			return -math.Log(1-rand.Float64()) / *vc.Rate
		}, nil

	default:
		return nil, errors.New("unknown distribution type")
	}
}

type VariableCostResult struct {
	Total   float64
	PerUser float64
	ByLayer types.CostLayerBreakdown
	ByCost  []VariableCostLineItem
	Samples []float64
}

type VariableCostLineItem struct {
	Name         string
	CostPerUnit  float64
	UnitsPerUser float64
	Total        float64
	Layer        types.CostLayer
	Distribution types.DistributionType
}

func CalculateVariableCosts(variableCosts []types.VariableCost, users int) VariableCostResult {
	result := VariableCostResult{
		ByLayer: types.CostLayerBreakdown{
			Infrastructure: 0,
			Application:    0,
			Service:        0,
		},
		ByCost:  make([]VariableCostLineItem, 0, len(variableCosts)),
		Samples: make([]float64, 0),
	}

	if users == 0 {
		return result
	}

	for _, vc := range variableCosts {
		baseUnits := vc.UnitsPerUser * float64(users)
		baseCost := vc.CostPerUnit * baseUnits

		lineItem := VariableCostLineItem{
			Name:         vc.Name,
			CostPerUnit:  vc.CostPerUnit,
			UnitsPerUser: vc.UnitsPerUser,
			Total:        baseCost,
			Layer:        vc.Layer,
			Distribution: vc.Distribution,
		}

		result.Total += baseCost
		result.ByCost = append(result.ByCost, lineItem)

		switch vc.Layer {
		case types.LayerInfrastructure:
			result.ByLayer.Infrastructure += baseCost
		case types.LayerApplication:
			result.ByLayer.Application += baseCost
		case types.LayerService:
			result.ByLayer.Service += baseCost
		}
	}

	result.PerUser = result.Total / float64(users)

	result.Total = math.Round(result.Total*100) / 100
	result.PerUser = math.Round(result.PerUser*100) / 100
	result.ByLayer.Infrastructure = math.Round(result.ByLayer.Infrastructure*100) / 100
	result.ByLayer.Application = math.Round(result.ByLayer.Application*100) / 100
	result.ByLayer.Service = math.Round(result.ByLayer.Service*100) / 100

	for i := range result.ByCost {
		result.ByCost[i].Total = math.Round(result.ByCost[i].Total*100) / 100
	}

	return result
}

func CalculateVariableCostWithVariability(variableCosts []types.VariableCost, users int) VariableCostResult {
	result := VariableCostResult{
		ByLayer: types.CostLayerBreakdown{
			Infrastructure: 0,
			Application:    0,
			Service:        0,
		},
		ByCost:  make([]VariableCostLineItem, 0, len(variableCosts)),
		Samples: make([]float64, 1),
	}

	if users == 0 {
		return result
	}

	for _, vc := range variableCosts {
		generator, err := NewDistributionGenerator(vc)
		if err != nil {
			baseCost := vc.CostPerUnit * vc.UnitsPerUser * float64(users)
			result.Total += baseCost

			switch vc.Layer {
			case types.LayerInfrastructure:
				result.ByLayer.Infrastructure += baseCost
			case types.LayerApplication:
				result.ByLayer.Application += baseCost
			case types.LayerService:
				result.ByLayer.Service += baseCost
			}
			continue
		}

		totalUnits := 0.0
		for i := 0; i < users; i++ {
			units := generator()
			if units < 0 {
				units = 0
			}
			totalUnits += units
		}

		variableCost := vc.CostPerUnit * totalUnits

		result.Total += variableCost

		lineItem := VariableCostLineItem{
			Name:         vc.Name,
			CostPerUnit:  vc.CostPerUnit,
			UnitsPerUser: totalUnits / float64(users),
			Total:        variableCost,
			Layer:        vc.Layer,
			Distribution: vc.Distribution,
		}
		result.ByCost = append(result.ByCost, lineItem)

		switch vc.Layer {
		case types.LayerInfrastructure:
			result.ByLayer.Infrastructure += variableCost
		case types.LayerApplication:
			result.ByLayer.Application += variableCost
		case types.LayerService:
			result.ByLayer.Service += variableCost
		}
	}

	result.PerUser = result.Total / float64(users)
	result.Samples[0] = result.Total

	result.Total = math.Round(result.Total*100) / 100
	result.PerUser = math.Round(result.PerUser*100) / 100
	result.ByLayer.Infrastructure = math.Round(result.ByLayer.Infrastructure*100) / 100
	result.ByLayer.Application = math.Round(result.ByLayer.Application*100) / 100
	result.ByLayer.Service = math.Round(result.ByLayer.Service*100) / 100

	for i := range result.ByCost {
		result.ByCost[i].Total = math.Round(result.ByCost[i].Total*100) / 100
		result.ByCost[i].UnitsPerUser = math.Round(result.ByCost[i].UnitsPerUser*100) / 100
	}

	return result
}
