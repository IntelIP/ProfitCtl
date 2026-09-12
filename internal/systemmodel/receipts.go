package systemmodel

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Receipt accepts only the documented sanitized UsageReceipt projection.
// Missing counters stay missing; raw transcript and arbitrary extra fields fail.
type Receipt struct {
	Provider          string                     `json:"provider"`
	Model             *string                    `json:"model,omitempty"`
	Region            *string                    `json:"region,omitempty"`
	InputTokens       *int64                     `json:"input_tokens,omitempty"`
	CachedInputTokens *int64                     `json:"cached_input_tokens,omitempty"`
	OutputTokens      *int64                     `json:"output_tokens,omitempty"`
	Cost              *float64                   `json:"cost_usd,omitempty"`
	Latency           *float64                   `json:"latency_ms,omitempty"`
	ToolCalls         *int64                     `json:"tool_calls,omitempty"`
	Metadata          map[string]json.RawMessage `json:"metadata,omitempty"`
}
type ReceiptExport struct {
	SchemaVersion string    `json:"schema_version"`
	AssessmentID  string    `json:"assessment_id"`
	Outcome       string    `json:"outcome"`
	Receipts      []Receipt `json:"receipts"`
}
type ReceiptGroup struct {
	CachedInputTokens *int64   `json:"cached_input_tokens"`
	LatencySum        *float64 `json:"sum_receipt_latency_ms"`
	Provider          string   `json:"provider"`
	Model             string   `json:"model"`
	Count             int      `json:"receipt_count"`
	KnownCost         float64  `json:"reported_cost_usd"`
	UnknownCosts      int      `json:"unknown_cost_receipts"`
	InputTokens       *int64   `json:"input_tokens"`
	OutputTokens      *int64   `json:"output_tokens"`
	ToolCalls         *int64   `json:"tool_calls"`
}
type ReceiptSummary struct {
	SchemaVersion string         `json:"schema_version"`
	SourceSHA256  string         `json:"source_sha256"`
	Outcome       string         `json:"outcome"`
	Groups        []ReceiptGroup `json:"groups"`
	Total         *float64       `json:"total_cost_usd"`
	KnownCost     float64        `json:"reported_cost_usd"`
	Limitations   []string       `json:"limitations"`
}

func SummarizeReceipts(input ReceiptExport, raw []byte) (ReceiptSummary, error) {
	if input.SchemaVersion != "condere.usage-export/v1" || !idPattern.MatchString(input.AssessmentID) || len(input.Receipts) == 0 || len(input.Receipts) > 2048 {
		return ReceiptSummary{}, fmt.Errorf("invalid receipt export schema, identity or count")
	}
	switch input.Outcome {
	case "succeeded", "failed", "cancelled", "unknown":
	default:
		return ReceiptSummary{}, fmt.Errorf("invalid assessment outcome")
	}
	result := ReceiptSummary{SchemaVersion: "profitctl.usage-summary/v1", SourceSHA256: fmt.Sprintf("%x", sha256.Sum256(raw)), Outcome: input.Outcome, Groups: []ReceiptGroup{}, Limitations: []string{"Observed sample only; not monthly workload or a success-rate estimate.", "Missing provider cost remains unknown. Do not add reported cost to token-price estimates.", "Raw metadata and assessment identifiers are not echoed."}}
	groups := map[string]*ReceiptGroup{}
	unknown := false
	for _, r := range input.Receipts {
		if !idPattern.MatchString(r.Provider) {
			return result, fmt.Errorf("invalid provider")
		}
		model := ""
		if r.Model != nil {
			model = *r.Model
		}
		if len(model) > 200 || strings.ContainsAny(model, "\r\n") {
			return result, fmt.Errorf("invalid model")
		}
		for _, n := range []*int64{r.InputTokens, r.CachedInputTokens, r.OutputTokens, r.ToolCalls} {
			if n != nil && (*n < 0 || *n > 1000000000000) {
				return result, fmt.Errorf("invalid usage counter")
			}
		}
		if r.Cost != nil && !finite(*r.Cost) {
			return result, fmt.Errorf("invalid receipt cost")
		}
		if r.Latency != nil && !finite(*r.Latency) {
			return result, fmt.Errorf("invalid latency")
		}
		status := ""
		if rawStatus, ok := r.Metadata["cost_status"]; ok {
			if err := json.Unmarshal(rawStatus, &status); err != nil {
				return result, fmt.Errorf("invalid cost status")
			}
		}
		switch status {
		case "", "reported", "not_reported", "unknown", "estimated":
		default:
			return result, fmt.Errorf("unsupported cost status")
		}
		uncertain := false
		if rawUncertain, ok := r.Metadata["unknown_spend"]; ok {
			if err := json.Unmarshal(rawUncertain, &uncertain); err != nil {
				return result, fmt.Errorf("invalid unknown_spend")
			}
		}
		known := r.Cost != nil && status == "reported" && !uncertain
		key := r.Provider + "\x00" + model
		g := groups[key]
		if g == nil {
			zero := int64(0)
			other := int64(0)
			third := int64(0)
			g = &ReceiptGroup{Provider: r.Provider, Model: model, InputTokens: &zero, OutputTokens: &other, ToolCalls: &third, CachedInputTokens: new(int64), LatencySum: number(0)}
			groups[key] = g
		}
		g.Count++
		if known {
			g.KnownCost += *r.Cost
		} else {
			g.UnknownCosts++
			unknown = true
		}
		for _, pair := range []struct {
			dst **int64
			src *int64
		}{{&g.InputTokens, r.InputTokens}, {&g.CachedInputTokens, r.CachedInputTokens}, {&g.OutputTokens, r.OutputTokens}, {&g.ToolCalls, r.ToolCalls}} {
			if pair.src == nil {
				*pair.dst = nil
			} else if *pair.dst != nil {
				**pair.dst += *pair.src
			}
		}
		if r.Latency == nil {
			g.LatencySum = nil
		} else if g.LatencySum != nil {
			*g.LatencySum += *r.Latency
		}
		if g.LatencySum != nil && !finite(*g.LatencySum) {
			return result, fmt.Errorf("receipt latency overflow")
		}
		if !finite(g.KnownCost) {
			return result, fmt.Errorf("receipt sum overflow")
		}
	}
	keys := []string{}
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		g := groups[key]
		result.Groups = append(result.Groups, *g)
		result.KnownCost += g.KnownCost
	}
	if !finite(result.KnownCost) {
		return result, fmt.Errorf("receipt total overflow")
	}
	if !unknown {
		result.Total = number(result.KnownCost)
	}
	return result, nil
}
