package plugin

import (
	"errors"
	"testing"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

func TestAPMPlugin_SetConfig(t *testing.T) {
	testCases := []struct {
		inputConfig  map[string]string
		expectOutput error
		name         string
	}{
		{
			inputConfig:  map[string]string{"wf_api_key": "123"},
			expectOutput: errors.New(`"address" config value cannot be empty`),
			name:         "no required address parameters set",
		},
		{
			inputConfig:  map[string]string{"address": "eample.com"},
			expectOutput: errors.New(`"wf_api_key" config value cannot be empty`),
			name:         "no required wf_api_key parameters set",
		},
		{
			inputConfig:  map[string]string{"address": "example.com", "wf_api_key": "123"},
			expectOutput: nil,
			name:         "required and valid config parameters set",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			apmPlugin := APMPlugin{logger: hclog.NewNullLogger()}

			actualOutput := apmPlugin.SetConfig(tc.inputConfig)
			assert.Equal(t, tc.expectOutput, actualOutput, tc.name)

			if actualOutput == nil {
				assert.NotNil(t, apmPlugin.client)
			}
		})
	}
}
