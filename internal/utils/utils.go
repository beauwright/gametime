package utils

// MapWithIndex takes a slice and a function to apply to each element, index is passed as the second argument of the function
func MapWithIndex[T, U any](ts []T, f func(T, int) U) []U {
	us := make([]U, len(ts))
	for i := range ts {
		us[i] = f(ts[i], i)
	}
	return us
}
