package util // not using util_test because I don't want none to be exposed outside the package

import (
	"reflect"
	"testing"
)

// TODO DRY
// TODO test with different sized inputs
func TestUnionOverlapping(t *testing.T) {
	expectedUnion := Set[int]{1: none{}, 2: none{}, 3: none{}, 4: none{}}
	s1 := Set[int]{1: none{}, 2: none{}, 3: none{}}
	s2 := Set[int]{2: none{}, 3: none{}, 4: none{}}

	union := Union(s1, s2)
	if !reflect.DeepEqual(union, expectedUnion) {
		t.Fatalf("union does not match expected values")
	}

	if !reflect.DeepEqual(union, expectedUnion) {
		t.Fatalf("union does not match expected values")
	}
}

func TestUnionSame(t *testing.T) {
	expectedUnion := Set[int]{1: none{}, 2: none{}, 3: none{}}
	s1 := Set[int]{1: none{}, 2: none{}, 3: none{}}
	s2 := Set[int]{1: none{}, 2: none{}, 3: none{}}

	union := Union(s1, s2)
	if !reflect.DeepEqual(union, expectedUnion) {
		t.Fatalf("union does not match expected values")
	}

	union = Union(s2, s1)
	if !reflect.DeepEqual(union, expectedUnion) {
		t.Fatalf("union does not match expected values")
	}
}

func TestUnionDisjoint(t *testing.T) {
	expectedUnion := Set[int]{1: none{}, 2: none{}, 3: none{}, 4: none{}}
	s1 := Set[int]{1: none{}, 2: none{}}
	s2 := Set[int]{3: none{}, 4: none{}}

	union := Union(s1, s2)
	if !reflect.DeepEqual(union, expectedUnion) {
		t.Fatalf("union does not match expected values")
	}

	union = Union(s2, s1)
	if !reflect.DeepEqual(union, expectedUnion) {
		t.Fatalf("union does not match expected values")
	}
}

func TestUnionOneEmpty(t *testing.T) {
	expectedUnion := Set[int]{1: none{}, 2: none{}}
	s1 := Set[int]{1: none{}, 2: none{}}
	s2 := Set[int]{}

	union := Union(s1, s2)
	if !reflect.DeepEqual(union, expectedUnion) {
		t.Fatalf("union does not match expected values")
	}

	union = Union(s2, s1)
	if !reflect.DeepEqual(union, expectedUnion) {
		t.Fatalf("union does not match expected values")
	}
}

func TestUnionBothEmpty(t *testing.T) {
	s1 := Set[int]{}
	s2 := Set[int]{}
	union := Union(s1, s2)

	if len(union) != 0 {
		t.Fatalf("union of empty sets has size > 0")
	}
}

func TestIntersectionOverlapping(t *testing.T) {
	expectedIntersection := Set[int]{2: none{}, 3: none{}}
	s1 := Set[int]{1: none{}, 2: none{}, 3: none{}}
	s2 := Set[int]{2: none{}, 3: none{}, 4: none{}}

	intersection := Intersection(s1, s2)
	if !reflect.DeepEqual(intersection, expectedIntersection) {
		t.Fatalf("intersection does not match expected values")
	}

	intersection = Intersection(s2, s1)
	if !reflect.DeepEqual(intersection, expectedIntersection) {
		t.Fatalf("intersection does not match expected values")
	}
}

func TestIntersectionSame(t *testing.T) {
	expectedIntersection := Set[int]{1: none{}, 2: none{}, 3: none{}}
	s1 := Set[int]{1: none{}, 2: none{}, 3: none{}}
	s2 := Set[int]{1: none{}, 2: none{}, 3: none{}}

	intersection := Intersection(s1, s2)
	if !reflect.DeepEqual(intersection, expectedIntersection) {
		t.Fatalf("intersection does not match expected values")
	}

	intersection = Intersection(s2, s1)
	if !reflect.DeepEqual(intersection, expectedIntersection) {
		t.Fatalf("intersection does not match expected values")
	}
}

func TestIntersectionDisjoint(t *testing.T) {
	s1 := Set[int]{1: none{}, 2: none{}}
	s2 := Set[int]{3: none{}, 4: none{}}

	intersection := Intersection(s1, s2)
	if len(intersection) != 0 {
		t.Fatalf("intersection on disjoint sets has size > 0")
	}

	intersection = Intersection(s2, s1)
	if len(intersection) != 0 {
		t.Fatalf("intersection on disjoint sets has size > 0")
	}
}

func TestIntersectiononeEmpty(t *testing.T) {
	s1 := Set[int]{1: none{}, 2: none{}}
	s2 := Set[int]{}
	intersection := Intersection(s1, s2)
	if len(intersection) != 0 {
		t.Fatalf("intersection with 1 empty set has size > 0")
	}

	intersection = Intersection(s2, s1)
	if len(intersection) != 0 {
		t.Fatalf("intersection with 1 empty set has size > 0")
	}
}

func TestIntersectionBothEmpty(t *testing.T) {
	s1 := Set[int]{}
	s2 := Set[int]{}
	intersection := Intersection(s1, s2)

	if len(intersection) != 0 {
		t.Fatalf("intersection on empty sets has size > 0")
	}
}

func TestDifferenceOverlapping(t *testing.T) {
	s1 := Set[int]{1: none{}, 2: none{}, 3: none{}}
	s2 := Set[int]{2: none{}, 3: none{}, 4: none{}}

	difference := Difference(s1, s2)
	expectedDifference := Set[int]{1: none{}}
	if !reflect.DeepEqual(difference, expectedDifference) {
		t.Fatalf("difference does not match expected values")
	}

	difference = Difference(s2, s1)
	expectedDifference = Set[int]{4: none{}}
	if !reflect.DeepEqual(difference, expectedDifference) {
		t.Fatalf("difference does not match expected values")
	}
}

func TestDifferenceSame(t *testing.T) {
	s1 := Set[int]{1: none{}, 2: none{}}
	s2 := Set[int]{1: none{}, 2: none{}}

	difference := Difference(s1, s2)
	if len(difference) != 0 {
		t.Fatalf("difference of identical sets is not empty")
	}
}

func TestDifferenceDisjoint(t *testing.T) {
	s1 := Set[int]{1: none{}, 2: none{}}
	s2 := Set[int]{3: none{}, 4: none{}}

	difference := Difference(s1, s2)
	expectedDifference := Set[int]{1: none{}, 2: none{}}
	if !reflect.DeepEqual(difference, expectedDifference) {
		t.Fatalf("difference does not match expected values")
	}

	difference = Difference(s2, s1)
	expectedDifference = Set[int]{3: none{}, 4: none{}}
	if !reflect.DeepEqual(difference, expectedDifference) {
		t.Fatalf("difference does not match expected values")
	}
}

func TestDifferenceOneEmpty(t *testing.T) {
	s1 := Set[int]{1: none{}, 2: none{}, 3: none{}}
	s2 := Set[int]{}

	difference := Difference(s1, s2)
	expectedDifference := Set[int]{1: none{}, 2: none{}, 3: none{}}
	if !reflect.DeepEqual(difference, expectedDifference) {
		t.Fatalf("difference does not match expected values")
	}

	difference = Difference(s2, s1)
	expectedDifference = Set[int]{}
	if !reflect.DeepEqual(difference, expectedDifference) {
		t.Fatalf("difference does not match expected values")
	}
}

func TestDifferenceBothEmpty(t *testing.T) {
	s1 := Set[int]{}
	s2 := Set[int]{}

	difference := Difference(s1, s2)
	if len(difference) != 0 {
		t.Fatalf("difference on empty sets has size > 0")
	}
}
