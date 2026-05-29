package types

import "math"

// FixedCost represents a fixed cost with amount and billing period
type FixedCost struct {
	Name           string               `yaml:"name" validate:"required"`
	Amount         float64              `yaml:"amount" validate:"gte=0"`
	Period         CostPeriod           `yaml:"period" validate:"required,oneof=monthly yearly daily"`
	Layer          CostLayer            `yaml:"layer" validate:"required,oneof=infrastructure application service"`
	EconomicsLayer EconomicsLayer       `yaml:"economics_layer,omitempty" validate:"omitempty,oneof=delivery productization adoption"`
	Allocation     *EconomicsAllocation `yaml:"allocation,omitempty"`
	Source         *CostSource          `yaml:"source,omitempty"`
}

// VariableCost represents a per-user variable cost with statistical distribution
type VariableCost struct {
	Name           string                `yaml:"name" validate:"required"`
	CostPerUnit    float64               `yaml:"cost_per_unit" validate:"gte=0"`
	UnitsPerUser   float64               `yaml:"units_per_user" validate:"gte=0"`
	Distribution   DistributionType      `yaml:"distribution" validate:"required,oneof=normal exponential uniform"`
	EconomicsLayer EconomicsLayer        `yaml:"economics_layer,omitempty" validate:"omitempty,oneof=delivery productization adoption"`
	Allocation     *EconomicsAllocation  `yaml:"allocation,omitempty"`
	UserScope      VariableCostUserScope `yaml:"user_scope,omitempty" validate:"omitempty,oneof=all_users paid_users"`

	// Distribution-specific parameters (validated based on Distribution field)
	Mean   *float64 `yaml:"mean,omitempty"`   // Required for normal
	StdDev *float64 `yaml:"stddev,omitempty"` // Required for normal
	Min    *float64 `yaml:"min,omitempty"`    // Required for uniform
	Max    *float64 `yaml:"max,omitempty"`    // Required for uniform
	Rate   *float64 `yaml:"rate,omitempty"`   // Required for exponential

	Layer  CostLayer   `yaml:"layer" validate:"required,oneof=infrastructure application service"`
	Source *CostSource `yaml:"source,omitempty"`
}

// EconomicsAllocation defines how non-delivery costs should be allocated in future reporting layers.
type EconomicsAllocation struct {
	Mode    EconomicsAllocationMode `yaml:"mode" validate:"required,oneof=fixed_monthly per_active_user per_design_partner per_workspace per_release"`
	Divisor int                     `yaml:"divisor,omitempty" validate:"omitempty,min=1"`
}

// CostSource captures provenance for cost and usage assumptions.
type CostSource struct {
	Type       CostSourceType       `yaml:"type" validate:"required,oneof=template user_supplied repo_detected telemetry invoice provider_catalog"`
	Confidence CostSourceConfidence `yaml:"confidence" validate:"required,oneof=low medium high"`
	URL        string               `yaml:"url,omitempty" validate:"omitempty,url"`
	CapturedAt string               `yaml:"captured_at,omitempty"`
	Note       string               `yaml:"note,omitempty"`
}

// CostLayerBreakdown separates costs by infrastructure layer
type CostLayerBreakdown struct {
	Infrastructure float64 `json:"infrastructure"`
	Application    float64 `json:"application"`
	Service        float64 `json:"service"`
}

// EconomicsLayerBreakdown separates costs by business economics bucket.
type EconomicsLayerBreakdown struct {
	Delivery       float64 `json:"delivery"`
	Productization float64 `json:"productization"`
	Adoption       float64 `json:"adoption"`
}

// Add places the amount into the normalized economics bucket.
func (b *EconomicsLayerBreakdown) Add(layer EconomicsLayer, amount float64) {
	switch NormalizeEconomicsLayer(layer) {
	case EconomicsLayerProductization:
		b.Productization += amount
	case EconomicsLayerAdoption:
		b.Adoption += amount
	default:
		b.Delivery += amount
	}
}

// Round keeps all economics totals at financial precision.
func (b *EconomicsLayerBreakdown) Round() {
	b.Delivery = math.Round(b.Delivery*100) / 100
	b.Productization = math.Round(b.Productization*100) / 100
	b.Adoption = math.Round(b.Adoption*100) / 100
}
