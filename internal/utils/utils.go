package utils

func Map[T, U any](ts []T, f func(T) U) []U {
	us := make([]U, len(ts))
	for i := range ts {
		us[i] = f(ts[i])
	}
	return us
}

// MapWithIndex takes a slice and a function to apply to each element, index is passed as the second argument of the function
func MapWithIndex[T, U any](ts []T, f func(T, int) U) []U {
	us := make([]U, len(ts))
	for i := range ts {
		us[i] = f(ts[i], i)
	}
	return us
}

func Filter[T any](in []T, f func(T) bool) []T {
	result := make([]T, 0)
	for i := range in {
	   if f(in[i]) {
               result = append(result, in[i])
           }
	}
	return result
}

