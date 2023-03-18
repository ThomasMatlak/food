package util

func UnpackArray[T any](arr []any) []T {
	r := make([]T, len(arr))
	for i, e := range arr {
		r[i] = e.(T)
	}
	return r
}
