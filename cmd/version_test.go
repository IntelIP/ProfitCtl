package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetVersion(t *testing.T) {
	oldVersion := buildVersion
	defer SetVersion(oldVersion)

	SetVersion("v1.2.3")
	assert.Equal(t, "v1.2.3", buildVersion)
	assert.Equal(t, "v1.2.3", rootCmd.Version)

	SetVersion(" ")
	assert.Equal(t, "dev", buildVersion)
	assert.Equal(t, "dev", rootCmd.Version)
}

func TestVersionCommand(t *testing.T) {
	oldVersion := buildVersion
	defer SetVersion(oldVersion)
	SetVersion("v1.2.3")

	var output bytes.Buffer
	versionCmd.SetOut(&output)
	defer versionCmd.SetOut(nil)

	versionCmd.Run(versionCmd, nil)
	assert.Equal(t, "profitctl v1.2.3\n", output.String())
}
