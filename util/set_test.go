package util

import "testing"

// TODO DRY
// TODO test with different sized inputs
func TestUnionOverlapping(t *testing.T) {
	expectedUnion := Set[int]{1: true, 2: true, 3: true, 4: true}
	s1 := Set[int]{1: true, 2: true, 3: true}
	s2 := Set[int]{2: true, 3: true, 4: true}

	union := Union(s1, s2)
	if len(union) < len(expectedUnion) {
		t.Fatalf("union is too small")
	}
	if len(union) > len(expectedUnion) {
		t.Fatalf("union is too big")
	}
	for k := range expectedUnion {
		if !union[k] {
			t.Fatalf("union does not match expected values")
		}
	}

	union = Union(s2, s1)
	if len(union) < len(expectedUnion) {
		t.Fatalf("union is too small")
	}
	if len(union) > len(expectedUnion) {
		t.Fatalf("union is too big")
	}
	for k := range expectedUnion {
		if !union[k] {
			t.Fatalf("union does not match expected values")
		}
	}
}

func TestUnionSame(t *testing.T) {
	expectedUnion := Set[int]{1: true, 2: true, 3: true}
	s1 := Set[int]{1: true, 2: true, 3: true}
	s2 := Set[int]{1: true, 2: true, 3: true}

	union := Union(s1, s2)
	if len(union) < len(expectedUnion) {
		t.Fatalf("union is too small")
	}
	if len(union) > len(expectedUnion) {
		t.Fatalf("union is too big")
	}
	for k := range expectedUnion {
		if !union[k] {
			t.Fatalf("union does not match expected values")
		}
	}

	union = Union(s2, s1)
	if len(union) < len(expectedUnion) {
		t.Fatalf("union is too small")
	}
	if len(union) > len(expectedUnion) {
		t.Fatalf("union is too big")
	}
	for k := range expectedUnion {
		if !union[k] {
			t.Fatalf("union does not match expected values")
		}
	}
}

func TestUnionDisjoint(t *testing.T) {
	expectedUnion := Set[int]{1: true, 2: true, 3: true, 4: true}
	s1 := Set[int]{1: true, 2: true}
	s2 := Set[int]{3: true, 4: true}

	union := Union(s1, s2)
	if len(union) < len(expectedUnion) {
		t.Fatalf("union is too small")
	}
	if len(union) > len(expectedUnion) {
		t.Fatalf("union is too big")
	}
	for k := range expectedUnion {
		if !union[k] {
			t.Fatalf("union does not match expected values")
		}
	}

	union = Union(s2, s1)
	if len(union) < len(expectedUnion) {
		t.Fatalf("union is too small")
	}
	if len(union) > len(expectedUnion) {
		t.Fatalf("union is too big")
	}
	for k := range expectedUnion {
		if !union[k] {
			t.Fatalf("union does not match expected values")
		}
	}
}

func TestUnionOneEmpty(t *testing.T) {
	expectedUnion := Set[int]{1: true, 2: true}
	s1 := Set[int]{1: true, 2: true}
	s2 := Set[int]{}

	union := Union(s1, s2)
	if len(union) < len(expectedUnion) {
		t.Fatalf("union is too small")
	}
	if len(union) > len(expectedUnion) {
		t.Fatalf("union is too big")
	}
	for k := range expectedUnion {
		if !union[k] {
			t.Fatalf("union does not match expected values")
		}
	}

	union = Union(s2, s1)
	if len(union) < len(expectedUnion) {
		t.Fatalf("union is too small")
	}
	if len(union) > len(expectedUnion) {
		t.Fatalf("union is too big")
	}
	for k := range expectedUnion {
		if !union[k] {
			t.Fatalf("union does not match expected values")
		}
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
	expectedIntersection := Set[int]{2: true, 3: true}
	s1 := Set[int]{1: true, 2: true, 3: true}
	s2 := Set[int]{2: true, 3: true, 4: true}

	intersection := Intersection(s1, s2)
	if len(intersection) < len(expectedIntersection) {
		t.Fatalf("intersection is too small")
	}
	if len(intersection) > len(expectedIntersection) {
		t.Fatalf("intersection is too big")
	}
	for k := range expectedIntersection {
		if !intersection[k] {
			t.Fatalf("intersection does not match expected values")
		}
	}

	intersection = Intersection(s2, s1)
	if len(intersection) < len(expectedIntersection) {
		t.Fatalf("intersection is too small")
	}
	if len(intersection) > len(expectedIntersection) {
		t.Fatalf("intersection is too big")
	}
	for k := range expectedIntersection {
		if !intersection[k] {
			t.Fatalf("intersection does not match expected values")
		}
	}
}

func TestIntersectionSame(t *testing.T) {
	expectedIntersection := Set[int]{1: true, 2: true, 3: true}
	s1 := Set[int]{1: true, 2: true, 3: true}
	s2 := Set[int]{1: true, 2: true, 3: true}

	intersection := Intersection(s1, s2)
	if len(intersection) < len(expectedIntersection) {
		t.Fatalf("intersection is too small")
	}
	if len(intersection) > len(expectedIntersection) {
		t.Fatalf("intersection is too big")
	}
	for k := range expectedIntersection {
		if !intersection[k] {
			t.Fatalf("intersection does not match expected values")
		}
	}

	intersection = Intersection(s2, s1)
	if len(intersection) < len(expectedIntersection) {
		t.Fatalf("intersection is too small")
	}
	if len(intersection) > len(expectedIntersection) {
		t.Fatalf("intersection is too big")
	}
	for k := range expectedIntersection {
		if !intersection[k] {
			t.Fatalf("intersection does not match expected values")
		}
	}
}

func TestIntersectionDisjoint(t *testing.T) {
	s1 := Set[int]{1: true, 2: true}
	s2 := Set[int]{3: true, 4: true}

	intersection := Intersection(s1, s2)
	if len(intersection) != 0 {
		t.Fatalf("intersection on disjoint sets has size > 0")
	}

	intersection = Intersection(s2, s1)
	if len(intersection) != 0 {
		t.Fatalf("intersection on disjoint sets has size > 0")
	}
}

func TestIntersectionOneEmpty(t *testing.T) {
	s1 := Set[int]{1: true, 2: true}
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
	s1 := Set[int]{1: true, 2: true, 3: true}
	s2 := Set[int]{2: true, 3: true, 4: true}

	difference := Difference(s1, s2)
	expectedDifference := Set[int]{1: true}
	if len(difference) < len(expectedDifference) {
		t.Fatalf("difference is too small")
	}
	if len(difference) > len(expectedDifference) {
		t.Fatalf("difference is too big")
	}
	for k := range expectedDifference {
		if !difference[k] {
			t.Fatalf("difference does not match expected values")
		}
	}

	difference = Difference(s2, s1)
	expectedDifference = Set[int]{4: true}
	if len(difference) < len(expectedDifference) {
		t.Fatalf("difference is too small")
	}
	if len(difference) > len(expectedDifference) {
		t.Fatalf("difference is too big")
	}
	for k := range expectedDifference {
		if !difference[k] {
			t.Fatalf("difference does not match expected values")
		}
	}
}

func TestDifferenceSame(t *testing.T) {
	s1 := Set[int]{1: true, 2: true}
	s2 := Set[int]{1: true, 2: true}

	difference := Difference(s1, s2)
	if len(difference) != 0 {
		t.Fatalf("difference of identical sets is not empty")
	}
}

func TestDifferenceDisjoint(t *testing.T) {
	s1 := Set[int]{1: true, 2: true}
	s2 := Set[int]{3: true, 4: true}

	difference := Difference(s1, s2)
	expectedDifference := Set[int]{1: true, 2: true}
	if len(difference) < len(expectedDifference) {
		t.Fatalf("difference is too small")
	}
	if len(difference) > len(expectedDifference) {
		t.Fatalf("difference is too big")
	}
	for k := range expectedDifference {
		if !difference[k] {
			t.Fatalf("difference does not match expected values")
		}
	}

	difference = Difference(s2, s1)
	expectedDifference = Set[int]{3: true, 4: true}
	if len(difference) < len(expectedDifference) {
		t.Fatalf("difference is too small")
	}
	if len(difference) > len(expectedDifference) {
		t.Fatalf("difference is too big")
	}
	for k := range expectedDifference {
		if !difference[k] {
			t.Fatalf("difference does not match expected values")
		}
	}
}

func TestDifferenceOneEmpty(t *testing.T) {
	s1 := Set[int]{1: true, 2: true, 3: true}
	s2 := Set[int]{}

	difference := Difference(s1, s2)
	expectedDifference := Set[int]{1: true, 2: true, 3: true}
	if len(difference) < len(expectedDifference) {
		t.Fatalf("difference is too small")
	}
	if len(difference) > len(expectedDifference) {
		t.Fatalf("difference is too big")
	}
	for k := range expectedDifference {
		if !difference[k] {
			t.Fatalf("difference does not match expected values")
		}
	}

	difference = Difference(s2, s1)
	expectedDifference = Set[int]{}
	if len(difference) < len(expectedDifference) {
		t.Fatalf("difference is too small")
	}
	if len(difference) > len(expectedDifference) {
		t.Fatalf("difference is too big")
	}
	for k := range expectedDifference {
		if !difference[k] {
			t.Fatalf("difference does not match expected values")
		}
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
