package util_test

import (
	"strings"
	"testing"

	"github.com/ThomasMatlak/food/util"
	"github.com/stretchr/testify/assert"
)

// TODO do generics allow these tests to be nicely put into table-driven sub-tests?

func TestMapAddToInt(t *testing.T) {
	f := func(x int) int { return x + 1 }

	input := []int{1, 2, 3}
	expectedOutput := []int{2, 3, 4}

	output := util.MapArray(input, f)

	assert.Equal(t, expectedOutput, output)
}

func TestMapExtractStructField(t *testing.T) {
	type foo struct {
		bar int
		baz int
	}

	f := func(x foo) int { return x.bar }

	input := []foo{{bar: 1, baz: 2}, {bar: 3, baz: 4}, {bar: 5, baz: 6}}
	expectedOutput := []int{1, 3, 5}

	output := util.MapArray(input, f)

	assert.Equal(t, expectedOutput, output)
}

func TestFilterArrayInt(t *testing.T) {
	f := func(x int) bool {
		return x <= 5
	}

	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	expectedOutput := []int{1, 2, 3, 4, 5}

	output := util.FilterArray(input, f)
	assert.Equal(t, expectedOutput, output)
}

func TestFilterArrayString(t *testing.T) {
	f := func(s string) bool {
		return strings.Contains(s, "test")
	}

	input := []string{"test value", "value", "t est", "value test"}
	expectedOutput := []string{"test value", "value test"}

	output := util.FilterArray(input, f)
	assert.Equal(t, expectedOutput, output)
}
