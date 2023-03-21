package util_test

import (
	"reflect"
	"testing"

	"github.com/ThomasMatlak/food/util"
)

func TestMapAddToInt(t *testing.T) {
	f := func(x int) int { return x + 1 }

	input := []int{1, 2, 3}
	expectedOutput := []int{2, 3, 4}

	output := util.MapArray(input, f)

	if !reflect.DeepEqual(output, expectedOutput) {
		t.Fatalf("mapped output does not match expected values")
	}
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

	if !reflect.DeepEqual(output, expectedOutput) {
		t.Fatalf("mapped output does not match expected values")
	}
}
