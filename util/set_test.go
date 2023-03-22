package util // not using util_test because I don't want none to be exposed outside the package

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO test with different sized inputs
// TODO test with one set being a subset of the other

type setOperationTestCase struct {
	name     string
	s1       Set[int]
	s2       Set[int]
	expected Set[int]
}

func TestUnion(t *testing.T) {
	testCases := []setOperationTestCase{
		{
			name:     "overlapping sets",
			s1:       Set[int]{1: none{}, 2: none{}, 3: none{}},
			s2:       Set[int]{2: none{}, 3: none{}, 4: none{}},
			expected: Set[int]{1: none{}, 2: none{}, 3: none{}, 4: none{}},
		},
		{
			name:     "same sets",
			s1:       Set[int]{1: none{}, 2: none{}, 3: none{}},
			s2:       Set[int]{1: none{}, 2: none{}, 3: none{}},
			expected: Set[int]{1: none{}, 2: none{}, 3: none{}},
		},
		{
			name:     "disjoint sets",
			s1:       Set[int]{1: none{}, 2: none{}},
			s2:       Set[int]{3: none{}, 4: none{}},
			expected: Set[int]{1: none{}, 2: none{}, 3: none{}, 4: none{}},
		},
		{
			name:     "one empty set",
			s1:       Set[int]{1: none{}, 2: none{}},
			s2:       Set[int]{},
			expected: Set[int]{1: none{}, 2: none{}},
		},
		{
			name:     "both empty sets",
			s1:       Set[int]{},
			s2:       Set[int]{},
			expected: Set[int]{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := Union(tc.s1, tc.s2)
			assert.Equal(t, actual, tc.expected)
			// union is commutative
			actual = Union(tc.s2, tc.s1)
			assert.Equal(t, actual, tc.expected)
		})
	}
}

func TestIntersection(t *testing.T) {
	testCases := []setOperationTestCase{
		{
			name:     "overlapping sets",
			s1:       Set[int]{1: none{}, 2: none{}, 3: none{}},
			s2:       Set[int]{2: none{}, 3: none{}, 4: none{}},
			expected: Set[int]{2: none{}, 3: none{}},
		},
		{
			name:     "same sets",
			s1:       Set[int]{1: none{}, 2: none{}, 3: none{}},
			s2:       Set[int]{1: none{}, 2: none{}, 3: none{}},
			expected: Set[int]{1: none{}, 2: none{}, 3: none{}},
		},
		{
			name:     "disjoint sets",
			s1:       Set[int]{1: none{}, 2: none{}},
			s2:       Set[int]{3: none{}, 4: none{}},
			expected: Set[int]{},
		},
		{
			name:     "one empty set",
			s1:       Set[int]{1: none{}, 2: none{}},
			s2:       Set[int]{},
			expected: Set[int]{},
		},
		{
			name:     "both empty sets",
			s1:       Set[int]{},
			s2:       Set[int]{},
			expected: Set[int]{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := Intersection(tc.s1, tc.s2)
			assert.Equal(t, actual, tc.expected)
			// intersection is commutative
			actual = Intersection(tc.s2, tc.s1)
			assert.Equal(t, actual, tc.expected)
		})
	}
}

func TestDifference(t *testing.T) {
	testCases := []setOperationTestCase{
		{
			name:     "overlapping sets",
			s1:       Set[int]{1: none{}, 2: none{}, 3: none{}},
			s2:       Set[int]{2: none{}, 3: none{}, 4: none{}},
			expected: Set[int]{1: none{}},
		},
		{
			name:     "overlapping sets (reversed)",
			s1:       Set[int]{2: none{}, 3: none{}, 4: none{}},
			s2:       Set[int]{1: none{}, 2: none{}, 3: none{}},
			expected: Set[int]{4: none{}},
		},
		{
			name:     "same sets",
			s1:       Set[int]{1: none{}, 2: none{}, 3: none{}},
			s2:       Set[int]{1: none{}, 2: none{}, 3: none{}},
			expected: Set[int]{},
		},
		{
			name:     "disjoint sets",
			s1:       Set[int]{1: none{}, 2: none{}},
			s2:       Set[int]{3: none{}, 4: none{}},
			expected: Set[int]{1: none{}, 2: none{}},
		},
		{
			name:     "disjoint sets (reversed)",
			s1:       Set[int]{3: none{}, 4: none{}},
			s2:       Set[int]{1: none{}, 2: none{}},
			expected: Set[int]{3: none{}, 4: none{}},
		},
		{
			name:     "one empty set",
			s1:       Set[int]{1: none{}, 2: none{}},
			s2:       Set[int]{},
			expected: Set[int]{1: none{}, 2: none{}},
		},
		{
			name:     "one empty set(reversed)",
			s1:       Set[int]{},
			s2:       Set[int]{1: none{}, 2: none{}},
			expected: Set[int]{},
		},
		{
			name:     "both empty sets",
			s1:       Set[int]{},
			s2:       Set[int]{},
			expected: Set[int]{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := Difference(tc.s1, tc.s2)
			assert.Equal(t, actual, tc.expected)
			// difference is not commutative
		})
	}
}
