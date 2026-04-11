package output

func recurringRevenueBase(result SimulationResult) float64 {
	if result.Revenue.RecurringTotal > 0 {
		return result.Revenue.RecurringTotal
	}
	return result.Revenue.Total
}

func operatingPaymentFeeAmount(result SimulationResult) float64 {
	if result.PaymentFees.OperatingAmount > 0 || result.PaymentFees.OneTimeAmount > 0 {
		return result.PaymentFees.OperatingAmount
	}
	return result.PaymentFees.Total
}

func bookedStressMargin(result SimulationResult, costPerUser float64) float64 {
	if result.Revenue.Total <= 0 || costPerUser <= 0 {
		return 0
	}
	totalCost := costPerUser*float64(result.Users) + result.PaymentFees.Total
	return ((result.Revenue.Total - totalCost) / result.Revenue.Total) * 100
}

func operatingStressMargin(result SimulationResult, costPerUser float64) float64 {
	revenueBase := recurringRevenueBase(result)
	if revenueBase <= 0 || costPerUser <= 0 {
		return 0
	}
	totalCost := costPerUser*float64(result.Users) + operatingPaymentFeeAmount(result)
	return ((revenueBase - totalCost) / revenueBase) * 100
}

func hasOperatingView(result SimulationResult) bool {
	return result.Revenue.OneTimeTotal > 0 || result.PaymentFees.OneTimeAmount > 0
}
