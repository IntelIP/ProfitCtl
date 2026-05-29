package types

// CostLayer represents the cost layer categorization
type CostLayer string

const (
	LayerInfrastructure CostLayer = "infrastructure"
	LayerApplication    CostLayer = "application"
	LayerService        CostLayer = "service"
)

// EconomicsLayer represents the business-level purpose of a cost.
type EconomicsLayer string

const (
	EconomicsLayerDelivery       EconomicsLayer = "delivery"
	EconomicsLayerProductization EconomicsLayer = "productization"
	EconomicsLayerAdoption       EconomicsLayer = "adoption"
)

// NormalizeEconomicsLayer preserves backward compatibility by treating omitted values as delivery costs.
func NormalizeEconomicsLayer(layer EconomicsLayer) EconomicsLayer {
	if layer == "" {
		return EconomicsLayerDelivery
	}

	return layer
}

// CostSourceType records where a scenario assumption came from.
type CostSourceType string

const (
	CostSourceTemplate        CostSourceType = "template"
	CostSourceUserSupplied    CostSourceType = "user_supplied"
	CostSourceRepoDetected    CostSourceType = "repo_detected"
	CostSourceTelemetry       CostSourceType = "telemetry"
	CostSourceInvoice         CostSourceType = "invoice"
	CostSourceProviderCatalog CostSourceType = "provider_catalog"
)

// CostSourceConfidence describes how much trust to place in a scenario assumption.
type CostSourceConfidence string

const (
	CostSourceConfidenceLow    CostSourceConfidence = "low"
	CostSourceConfidenceMedium CostSourceConfidence = "medium"
	CostSourceConfidenceHigh   CostSourceConfidence = "high"
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

// EconomicsAllocationMode describes how non-delivery costs should be allocated in later reporting layers.
type EconomicsAllocationMode string

const (
	AllocationFixedMonthly     EconomicsAllocationMode = "fixed_monthly"
	AllocationPerActiveUser    EconomicsAllocationMode = "per_active_user"
	AllocationPerDesignPartner EconomicsAllocationMode = "per_design_partner"
	AllocationPerWorkspace     EconomicsAllocationMode = "per_workspace"
	AllocationPerRelease       EconomicsAllocationMode = "per_release"
)

// VariableCostUserScope controls which users a variable cost applies to.
type VariableCostUserScope string

const (
	UserScopeAllUsers  VariableCostUserScope = "all_users"
	UserScopePaidUsers VariableCostUserScope = "paid_users"
)

// NormalizeVariableCostUserScope preserves backward compatibility by treating omitted values as all_users.
func NormalizeVariableCostUserScope(scope VariableCostUserScope) VariableCostUserScope {
	if scope == "" {
		return UserScopeAllUsers
	}

	return scope
}
