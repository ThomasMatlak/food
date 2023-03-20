package util

// TODO use map[T]struct{} for less memory
type Set[T comparable] map[T]bool

// TODO documentation
// TODO use pointers to Set?
func Union[T comparable](s1 Set[T], s2 Set[T]) Set[T] {
	union := Set[T]{}
	for k := range s1 {
		union[k] = true
	}
	for k := range s2 {
		union[k] = true
	}
	return union
}

func Intersection[T comparable](s1 Set[T], s2 Set[T]) Set[T] {
	intersection := Set[T]{}
	if len(s1) > len(s2) {
		s1, s2 = s2, s1
	}
	for k := range s1 {
		if s2[k] {
			intersection[k] = true
		}
	}
	return intersection
}

func Difference[T comparable](s1 Set[T], s2 Set[T]) Set[T] {
	difference := Set[T]{}
	for k := range s1 {
		if !s2[k] {
			difference[k] = true
		}
	}
	return difference
}

// TODO unit test
func AreDisjoint[T comparable](s1 Set[T], s2 Set[T]) bool {
	return len(Intersection(s1, s2)) == 0
}

// TODO unit test
func Complement[T comparable](universe Set[T], s Set[T]) Set[T] {
	return Difference(universe, s)
}

func ArrayToSet[T comparable](ts []T) Set[T] {
	res := Set[T]{}
	for i := range ts {
		res[ts[i]] = true
	}
	return res
}

func SetToArray[T comparable](s Set[T]) []T {
	keys := make([]T, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	return keys
}
