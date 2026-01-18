package pricing

import (
	"math"

	"github.com/profitctl/profitctl/pkg/types"
)

// MarginResult represents calculated margin and cost metrics
type MarginResult struct {
	GrossMargin       float64                // Gross margin percentage (0-100)
	MarginPercentage  float64                // Margin as decimal (0-1)
	CostPerUser       float64                // Total cost per user
	LayerMargins      LayerMarginBreakdown   // Margin contribution by layer
	LayerCostPercent  types.CostLayerBreakdown // Cost percentage by layer (0-100)
}

// LayerMarginBreakdown represents margin contribution by layer
type LayerMarginBreakdown struct {
	Infrastructure float64 // Infrastructure layer contribution to margin
	Application    float64 // Application layer contribution to margin
	Service        float64 // Service layer contribution to margin
}

// CalculateMargins calculates gross margin, cost per user, and layer breakdowns
// Returns margin result with all calculated metrics
func CalculateMargins(revenue float64, totalCosts float64, layerCosts types.CostLayerBreakdown, users int) MarginResult {
	result := MarginResult{}

	// Calculate gross margin percentage: (revenue - costs) / revenue * 100
	if revenue > 0 {
		result.MarginPercentage = (revenue - totalCosts) / revenue
		result.GrossMargin = result.MarginPercentage * 100
	} else {
		result.MarginPercentage = 0
		result.GrossMargin = 0
	}

	// Calculate cost per user
	if users > 0 {
		result.CostPerUser = totalCosts / float64(users)
	} else {
		result.CostPerUser = 0
	}

	// Calculate layer cost percentages (what % of total costs each layer represents)
	if totalCosts > 0 {
		result.LayerCostPercent.Infrastructure = (layerCosts.Infrastructure / totalCosts) * 100
		result.LayerCostPercent.Application = (layerCosts.Application / totalCosts) * 100
		result.LayerCostPercent.Service = (layerCosts.Service / totalCosts) * 100
	} else {
		result.LayerCostPercent = types.CostLayerBreakdown{}
	}

	// Calculate layer margin contribution (revenue - layer cost)
	// This shows how much each layer contributes to the margin
	if revenue > 0 {
		result.LayerMargins.Infrastructure = revenue - layerCosts.Infrastructure
		result.LayerMargins.Application = revenue - layerCosts.Application
		result.LayerMargins.Service = revenue - layerCosts.Service
	} else {
		result.LayerMargins = LayerMarginBreakdown{}
	}

	// Round all values to 2 decimal places for financial precision
	result.GrossMargin = math.Round(result.GrossMargin*100) / 100
	result.MarginPercentage = math.Round(result.MarginPercentage*10000) / 10000 // 4 decimals for percentage
	result.CostPerUser = math.Round(result.CostPerUser*100) / 100
	result.LayerMargins.Infrastructure = math.Round(result.LayerMargins.Infrastructure*100) / 100
	result.LayerMargins.Application = math.Round(result.LayerMargins.Application*100) / 100
	result.LayerMargins.Service = math.Round(result.LayerMargins.Service*100) / 100
	result.LayerCostPercent.Infrastructure = math.Round(result.LayerCostPercent.Infrastructure*100) / 100
	result.LayerCostPercent.Application = math.Round(result.LayerCostPercent.Application*100) / 100
	result.LayerCostPercent.Service = math.Round(result.LayerCostPercent.Service*100) / 100

	return result
}