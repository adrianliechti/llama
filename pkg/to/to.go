package to

func Ptr[T any](v T) *T {
	return &v
}

func Keys[M ~map[K]V, K comparable, V any](m M) []K {
	r := make([]K, 0, len(m))

	for k := range m {
		r = append(r, k)
	}

	return r
}

func Values[M ~map[K]V, K comparable, V any](m M) []V {
	r := make([]V, 0, len(m))

	for _, v := range m {
		r = append(r, v)
	}

	return r
}
