package types

// FixedCost represents a fixed cost with amount and billing period
type FixedCost struct {
	Name   string     `yaml:"name" validate:"required"`
	Amount float64    `yaml:"amount" validate:"required,min=0"`
	Period CostPeriod `yaml:"period" validate:"required,oneof=monthly yearly daily"`
	Layer  CostLayer  `yaml:"layer" validate:"required,oneof=infrastructure application service"`
}

// VariableCost represents a per-user variable cost with statistical distribution
type VariableCost struct {
	Name         string           `yaml:"name" validate:"required"`
	CostPerUnit  float64          `yaml:"cost_per_unit" validate:"required,min=0"`
	UnitsPerUser float64          `yaml:"units_per_user" validate:"required,min=0"`
	Distribution DistributionType `yaml:"distribution" validate:"required,oneof=normal exponential uniform"`

	// Distribution-specific parameters (validated based on Distribution field)
	Mean   *float64 `yaml:"mean,omitempty"`   // Required for normal
	StdDev *float64 `yaml:"stddev,omitempty"` // Required for normal
	Min    *float64 `yaml:"min,omitempty"`    // Required for uniform
	Max    *float64 `yaml:"max,omitempty"`    // Required for uniform
	Rate   *float64 `yaml:"rate,omitempty"`   // Required for exponential

	Layer CostLayer `yaml:"layer" validate:"required,oneof=infrastructure application service"`
}

// CostLayerBreakdown separates costs by infrastructure layer
type CostLayerBreakdown struct {
	Infrastructure float64 `json:"infrastructure"`
	Application    float64 `json:"application"`
	Service        float64 `json:"service"`
}
