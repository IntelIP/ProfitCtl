package llm

import "strings"

type trustedProviderSpec struct {
	OfficialDomain       string
	EvidenceMarkers      []string
	PricingPathFragments []string
}

var trustedProviderSpecs = map[string]trustedProviderSpec{
	"anthropic": {
		OfficialDomain:       "anthropic.com",
		EvidenceMarkers:      []string{"anthropic"},
		PricingPathFragments: []string{"/pricing"},
	},
	"aws": {
		OfficialDomain:       "aws.amazon.com",
		EvidenceMarkers:      []string{"amazonaws", "aws"},
		PricingPathFragments: []string{"/pricing", "/calculator"},
	},
	"azure": {
		OfficialDomain:       "azure.microsoft.com",
		EvidenceMarkers:      []string{"azure"},
		PricingPathFragments: []string{"/pricing"},
	},
	"buildkite": {
		OfficialDomain:       "buildkite.com",
		EvidenceMarkers:      []string{"buildkite"},
		PricingPathFragments: []string{"/pricing"},
	},
	"clerk": {
		OfficialDomain:       "clerk.com",
		EvidenceMarkers:      []string{"clerk"},
		PricingPathFragments: []string{"/pricing"},
	},
	"cloudflare": {
		OfficialDomain:       "cloudflare.com",
		EvidenceMarkers:      []string{"cloudflare"},
		PricingPathFragments: []string{"/pricing", "/plans"},
	},
	"datadog": {
		OfficialDomain:       "datadoghq.com",
		EvidenceMarkers:      []string{"datadog"},
		PricingPathFragments: []string{"/pricing"},
	},
	"exa": {
		OfficialDomain:       "exa.ai",
		EvidenceMarkers:      []string{"exa"},
		PricingPathFragments: []string{"/pricing"},
	},
	"gcp": {
		OfficialDomain:       "cloud.google.com",
		EvidenceMarkers:      []string{"cloud.google.com", "gcp", "google"},
		PricingPathFragments: []string{"/pricing", "/products/calculator"},
	},
	"github": {
		OfficialDomain:       "github.com",
		EvidenceMarkers:      []string{"github"},
		PricingPathFragments: []string{"/pricing"},
	},
	"google-cloud": {
		OfficialDomain:       "cloud.google.com",
		EvidenceMarkers:      []string{"cloud.google.com", "gcp", "google"},
		PricingPathFragments: []string{"/pricing", "/products/calculator"},
	},
	"mongodb-atlas": {
		OfficialDomain:       "mongodb.com",
		EvidenceMarkers:      []string{"mongodb"},
		PricingPathFragments: []string{"/pricing"},
	},
	"neon": {
		OfficialDomain:       "neon.com",
		EvidenceMarkers:      []string{"neon"},
		PricingPathFragments: []string{"/pricing"},
	},
	"openai": {
		OfficialDomain:       "openai.com",
		EvidenceMarkers:      []string{"openai"},
		PricingPathFragments: []string{"/pricing"},
	},
	"openrouter": {
		OfficialDomain:       "openrouter.ai",
		EvidenceMarkers:      []string{"openrouter"},
		PricingPathFragments: []string{"/pricing", "/models"},
	},
	"pinecone": {
		OfficialDomain:       "pinecone.io",
		EvidenceMarkers:      []string{"pinecone"},
		PricingPathFragments: []string{"/pricing"},
	},
	"resend": {
		OfficialDomain:       "resend.com",
		EvidenceMarkers:      []string{"resend"},
		PricingPathFragments: []string{"/pricing"},
	},
	"sentry": {
		OfficialDomain:       "sentry.io",
		EvidenceMarkers:      []string{"sentry"},
		PricingPathFragments: []string{"/pricing"},
	},
	"stripe": {
		OfficialDomain:       "stripe.com",
		EvidenceMarkers:      []string{"stripe"},
		PricingPathFragments: []string{"/pricing"},
	},
	"supabase": {
		OfficialDomain:       "supabase.com",
		EvidenceMarkers:      []string{"supabase"},
		PricingPathFragments: []string{"/pricing"},
	},
	"twilio": {
		OfficialDomain:       "twilio.com",
		EvidenceMarkers:      []string{"twilio"},
		PricingPathFragments: []string{"/pricing"},
	},
	"upstash": {
		OfficialDomain:       "upstash.com",
		EvidenceMarkers:      []string{"upstash"},
		PricingPathFragments: []string{"/pricing"},
	},
	"vercel": {
		OfficialDomain:       "vercel.com",
		EvidenceMarkers:      []string{"vercel"},
		PricingPathFragments: []string{"/pricing"},
	},
}

// TrustedOfficialDomain returns the code-owned official-domain binding for a
// supported provider. Model output can select a provider, but cannot invent or
// change the domain used for pricing research.
func TrustedOfficialDomain(provider string) (string, bool) {
	spec, ok := trustedProviderSpecs[normalizeProviderKey(provider)]
	return spec.OfficialDomain, ok
}

// TrustedPricingPathFragments returns official URL path fragments that can
// identify provider pricing pages without trusting search-result ranking.
func TrustedPricingPathFragments(provider string) ([]string, bool) {
	spec, ok := trustedProviderSpecs[normalizeProviderKey(provider)]
	if !ok {
		return nil, false
	}
	return append([]string(nil), spec.PricingPathFragments...), true
}

// ProviderEvidenceMatches binds a model-supplied repository excerpt to a
// deterministic marker for the selected provider.
func ProviderEvidenceMatches(provider, excerpt string) bool {
	spec, ok := trustedProviderSpecs[normalizeProviderKey(provider)]
	if !ok {
		return false
	}
	excerpt = strings.ToLower(excerpt)
	for _, marker := range spec.EvidenceMarkers {
		if containsEvidenceMarker(excerpt, strings.ToLower(marker)) {
			return true
		}
	}
	return false
}

func normalizeProviderKey(provider string) string {
	return strings.ToLower(strings.TrimSpace(provider))
}

func containsEvidenceMarker(excerpt, marker string) bool {
	for start := 0; start < len(excerpt); {
		index := strings.Index(excerpt[start:], marker)
		if index < 0 {
			return false
		}
		index += start
		beforeOK := index == 0 || !isASCIILetterOrDigit(excerpt[index-1])
		after := index + len(marker)
		afterOK := after == len(excerpt) || !isASCIILetterOrDigit(excerpt[after])
		if beforeOK && afterOK {
			return true
		}
		start = index + 1
	}
	return false
}

func isASCIILetterOrDigit(value byte) bool {
	return (value >= 'a' && value <= 'z') || (value >= '0' && value <= '9')
}
