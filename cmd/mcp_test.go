package cmd

import "testing"

func TestValidateLoopbackAddress(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		address string
		valid   bool
	}{
		{name: "ipv4", address: "127.0.0.1:8173", valid: true},
		{name: "ipv6", address: "[::1]:8173", valid: true},
		{name: "wildcard", address: "0.0.0.0:8173", valid: false},
		{name: "hostname", address: "localhost:8173", valid: false},
		{name: "missing port", address: "127.0.0.1", valid: false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateLoopbackAddress(testCase.address)
			if (err == nil) != testCase.valid {
				t.Fatalf("validateLoopbackAddress(%q) error = %v, valid = %t", testCase.address, err, testCase.valid)
			}
		})
	}
}
