package util

import "testing"

// TODO DRY
// TODO test with different sized inputs
func TestUnionOverlapping(t *testing.T) {
	expectedUnion := Set[int]{1: none{}, 2: none{}, 3: none{}, 4: none{}}
	s1 := Set[int]{1: none{}, 2: none{}, 3: none{}}
	s2 := Set[int]{2: none{}, 3: none{}, 4: none{}}

	union := Union(s1, s2)
	if len(union) < len(expectedUnion) {
		t.Fatalf("union is too small")
	}
	if len(union) > len(expectedUnion) {
		t.Fatalf("union is too big")
	}
	for k := range expectedUnion {
		if _, ok := union[k]; !ok {
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
		if _, ok := union[k]; !ok {
			t.Fatalf("union does not match expected values")
		}
	}
}

func TestUnionSame(t *testing.T) {
	expectedUnion := Set[int]{1: none{}, 2: none{}, 3: none{}}
	s1 := Set[int]{1: none{}, 2: none{}, 3: none{}}
	s2 := Set[int]{1: none{}, 2: none{}, 3: none{}}

	union := Union(s1, s2)
	if len(union) < len(expectedUnion) {
		t.Fatalf("union is too small")
	}
	if len(union) > len(expectedUnion) {
		t.Fatalf("union is too big")
	}
	for k := range expectedUnion {
		if _, ok := union[k]; !ok {
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
		if _, ok := union[k]; !ok {
			t.Fatalf("union does not match expected values")
		}
	}
}

func TestUnionDisjoint(t *testing.T) {
	expectedUnion := Set[int]{1: none{}, 2: none{}, 3: none{}, 4: none{}}
	s1 := Set[int]{1: none{}, 2: none{}}
	s2 := Set[int]{3: none{}, 4: none{}}

	union := Union(s1, s2)
	if len(union) < len(expectedUnion) {
		t.Fatalf("union is too small")
	}
	if len(union) > len(expectedUnion) {
		t.Fatalf("union is too big")
	}
	for k := range expectedUnion {
		if _, ok := union[k]; !ok {
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
		if _, ok := union[k]; !ok {
			t.Fatalf("union does not match expected values")
		}
	}
}

func TestUniononeEmpty(t *testing.T) {
	expectedUnion := Set[int]{1: none{}, 2: none{}}
	s1 := Set[int]{1: none{}, 2: none{}}
	s2 := Set[int]{}

	union := Union(s1, s2)
	if len(union) < len(expectedUnion) {
		t.Fatalf("union is too small")
	}
	if len(union) > len(expectedUnion) {
		t.Fatalf("union is too big")
	}
	for k := range expectedUnion {
		if _, ok := union[k]; !ok {
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
		if _, ok := union[k]; !ok {
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
	expectedIntersection := Set[int]{2: none{}, 3: none{}}
	s1 := Set[int]{1: none{}, 2: none{}, 3: none{}}
	s2 := Set[int]{2: none{}, 3: none{}, 4: none{}}

	intersection := Intersection(s1, s2)
	if len(intersection) < len(expectedIntersection) {
		t.Fatalf("intersection is too small")
	}
	if len(intersection) > len(expectedIntersection) {
		t.Fatalf("intersection is too big")
	}
	for k := range expectedIntersection {
		if _, ok := intersection[k]; !ok {
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
		if _, ok := intersection[k]; !ok {
			t.Fatalf("intersection does not match expected values")
		}
	}
}

func TestIntersectionSame(t *testing.T) {
	expectedIntersection := Set[int]{1: none{}, 2: none{}, 3: none{}}
	s1 := Set[int]{1: none{}, 2: none{}, 3: none{}}
	s2 := Set[int]{1: none{}, 2: none{}, 3: none{}}

	intersection := Intersection(s1, s2)
	if len(intersection) < len(expectedIntersection) {
		t.Fatalf("intersection is too small")
	}
	if len(intersection) > len(expectedIntersection) {
		t.Fatalf("intersection is too big")
	}
	for k := range expectedIntersection {
		if _, ok := intersection[k]; !ok {
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
		if _, ok := intersection[k]; !ok {
			t.Fatalf("intersection does not match expected values")
		}
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
	if len(difference) < len(expectedDifference) {
		t.Fatalf("difference is too small")
	}
	if len(difference) > len(expectedDifference) {
		t.Fatalf("difference is too big")
	}
	for k := range expectedDifference {
		if _, ok := difference[k]; !ok {
			t.Fatalf("difference does not match expected values")
		}
	}

	difference = Difference(s2, s1)
	expectedDifference = Set[int]{4: none{}}
	if len(difference) < len(expectedDifference) {
		t.Fatalf("difference is too small")
	}
	if len(difference) > len(expectedDifference) {
		t.Fatalf("difference is too big")
	}
	for k := range expectedDifference {
		if _, ok := difference[k]; !ok {
			t.Fatalf("difference does not match expected values")
		}
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
	if len(difference) < len(expectedDifference) {
		t.Fatalf("difference is too small")
	}
	if len(difference) > len(expectedDifference) {
		t.Fatalf("difference is too big")
	}
	for k := range expectedDifference {
		if _, ok := difference[k]; !ok {
			t.Fatalf("difference does not match expected values")
		}
	}

	difference = Difference(s2, s1)
	expectedDifference = Set[int]{3: none{}, 4: none{}}
	if len(difference) < len(expectedDifference) {
		t.Fatalf("difference is too small")
	}
	if len(difference) > len(expectedDifference) {
		t.Fatalf("difference is too big")
	}
	for k := range expectedDifference {
		if _, ok := difference[k]; !ok {
			t.Fatalf("difference does not match expected values")
		}
	}
}

func TestDifferenceOneEmpty(t *testing.T) {
	s1 := Set[int]{1: none{}, 2: none{}, 3: none{}}
	s2 := Set[int]{}

	difference := Difference(s1, s2)
	expectedDifference := Set[int]{1: none{}, 2: none{}, 3: none{}}
	if len(difference) < len(expectedDifference) {
		t.Fatalf("difference is too small")
	}
	if len(difference) > len(expectedDifference) {
		t.Fatalf("difference is too big")
	}
	for k := range expectedDifference {
		if _, ok := difference[k]; !ok {
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
		if _, ok := difference[k]; !ok {
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
