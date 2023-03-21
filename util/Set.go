package util

type none struct{}

type Set[T comparable] map[T]none

// TODO documentation
func Union[T comparable](s1 Set[T], s2 Set[T]) Set[T] {
	union := Set[T]{}
	for k := range s1 {
		union[k] = none{}
	}
	for k := range s2 {
		union[k] = none{}
	}
	return union
}

func Intersection[T comparable](s1 Set[T], s2 Set[T]) Set[T] {
	intersection := Set[T]{}
	if len(s1) > len(s2) {
		s1, s2 = s2, s1
	}
	for k := range s1 {
		if _, ok := s2[k]; ok {
			intersection[k] = none{}
		}
	}
	return intersection
}

func Difference[T comparable](s1 Set[T], s2 Set[T]) Set[T] {
	difference := Set[T]{}
	for k := range s1 {
		if _, ok := s2[k]; !ok {
			difference[k] = none{}
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
		res[ts[i]] = none{}
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
