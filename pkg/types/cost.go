package types

// CostLayer represents the cost layer categorization
type CostLayer string

const (
	LayerInfrastructure CostLayer = "infrastructure"
	LayerApplication    CostLayer = "application"
	LayerService        CostLayer = "service"
)

// CostPeriod represents the time period for fixed costs
type CostPeriod string

const (
	PeriodMonthly CostPeriod = "monthly"
	PeriodYearly  CostPeriod = "yearly"
	PeriodDaily   CostPeriod = "daily"
)

// DistributionType represents the statistical distribution type for variable costs
type DistributionType string

const (
	DistNormal      DistributionType = "normal"
	DistExponential DistributionType = "exponential"
	DistUniform     DistributionType = "uniform"
)
