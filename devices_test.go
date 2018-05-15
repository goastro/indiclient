package indiclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Groups(t *testing.T) {
	device := Device{
		Name: "TestDevice",
		TextProperties: map[string]TextProperty{
			"Prop1": {Group: "Group A"},
		},
		SwitchProperties: map[string]SwitchProperty{
			"Prop1": {Group: "Group A"},
		},
		NumberProperties: map[string]NumberProperty{
			"Prop1": {Group: "Group A"},
		},
		LightProperties: map[string]LightProperty{
			"Prop1": {Group: "Group B"},
		},
		BlobProperties: map[string]BlobProperty{
			"Prop1": {Group: "Group B"},
		},
	}

	groups := device.Groups()
	expected := []string{"Group A", "Group B"}

	require.NotNil(t, groups)
	assert.Equal(t, expected, groups)
}
