package settings

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseWideView(t *testing.T) {
	tests := []struct {
		value string
		want  int
	}{
		{"never", math.MaxInt},
		{"NEVER", math.MaxInt},
		{" never ", math.MaxInt},
		{"always", 0},
		{"Always", 0},
		{"240", 240},
		{" 120 ", 120},
		{"1", 1},
	}

	for _, tt := range tests {
		got, err := ParseWideView(tt.value)
		require.NoError(t, err, "value %q", tt.value)
		assert.Equal(t, tt.want, got, "value %q", tt.value)
	}
}

func TestParseWideView_Invalid(t *testing.T) {
	for _, value := range []string{"", "sometimes", "0", "-1", "24.5"} {
		_, err := ParseWideView(value)
		assert.Error(t, err, "value %q", value)
	}
}
