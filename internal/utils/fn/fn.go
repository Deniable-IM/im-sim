package fn

func Map[T, U any](slice []T, fn func(T) U) []U {
	res := make([]U, len(slice))
	for i, s := range slice {
		res[i] = fn(s)
	}
	return res
}
