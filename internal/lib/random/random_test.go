package random

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAlias_WithDefaultLength(t *testing.T) {
	tests := []struct {
		name           string
		expectedLength int
	}{
		{
			name:           "default length",
			expectedLength: 16,
		},
		{
			name:           "default length",
			expectedLength: 16,
		},
		{
			name:           "default length",
			expectedLength: 16,
		},
		{
			name:           "default length",
			expectedLength: 16,
		},
		{
			name:           "default length",
			expectedLength: 16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alias1 := Alias()
			alias2 := Alias()

			assert.Len(t, alias1, tt.expectedLength)
			assert.Len(t, alias2, tt.expectedLength)

			assert.NotEqual(t, alias1, alias2)
		})
	}

}

func TestAlias_WithSpecifiedLength(t *testing.T) {
	tests := []struct {
		name           string
		length         []int
		expectedLength int
	}{
		{
			name:           "len = 10",
			length:         []int{10},
			expectedLength: 10,
		},
		{
			name:           "len = 20",
			length:         []int{20},
			expectedLength: 20,
		},
		{
			name:           "negative len",
			length:         []int{-1},
			expectedLength: 0,
		},
		{
			name:           "more arguments",
			length:         []int{1, 2},
			expectedLength: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alias1 := Alias(tt.length...)
			alias2 := Alias(tt.length...)

			assert.Len(t, alias1, tt.expectedLength)
			assert.Len(t, alias2, tt.expectedLength)

			if alias1 != "" && alias2 != "" {
				assert.NotEqual(t, alias1, alias2)
			}
		})
	}
}
