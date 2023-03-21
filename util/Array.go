package util

func UnpackArray[T any](arr []any) []T {
	r := make([]T, len(arr))
	for i, e := range arr {
		r[i] = e.(T)
	}
	return r
}

func MapArray[T, V any](ts []T, fn func(T) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(t)
	}
	return result
}
