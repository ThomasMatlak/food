package model_test

import (
	"strings"
	"testing"

	"github.com/ThomasMatlak/food/model"
	"github.com/stretchr/testify/assert"
)

func TestResourceId(t *testing.T) {
	type testCase struct {
		name           string
		labels         []string
		shouldError    bool
		expectedPrefix string
	}

	testCases := []testCase{
		{
			name:           "Single label",
			labels:         []string{"Test"},
			expectedPrefix: "grn:tm-food:test:",
		},
		{
			name:           "Multiple labels",
			labels:         []string{"Label", "Test"},
			expectedPrefix: "grn:tm-food:label:test:",
		},
		{
			name:           "Labels should be sorted alphabetically",
			labels:         []string{"Test", "Resource", "Aardvark"},
			expectedPrefix: "grn:tm-food:aardvark:resource:test:",
		},
		{
			name:           "Empty strings should be filtered out",
			labels:         []string{"Test", "Resource", ""},
			expectedPrefix: "grn:tm-food:resource:test:",
		},
		{
			name:           "Empty lists should cause an error",
			labels:         []string{},
			shouldError:    true,
			expectedPrefix: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			id, err := model.ResourceId(tc.labels)
			if tc.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.True(t, strings.HasPrefix(id, tc.expectedPrefix))
		})
	}
}
