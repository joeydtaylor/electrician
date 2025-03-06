// functors.go file
package utils

// Map applies a function to each element in the slice.
func Map[T any](elems []T, f func(T) T) []T {
	result := make([]T, len(elems))
	for i, v := range elems {
		result[i] = f(v)
	}
	return result
}

// Filter returns a new slice holding only the elements of elems that satisfy f().
func Filter[T any](elems []T, f func(T) bool) []T {
	var result []T
	for _, v := range elems {
		if f(v) {
			result = append(result, v)
		}
	}
	return result
}

func Contains[T comparable](slice []T, element T) bool {
	for _, v := range slice {
		if v == element {
			return true
		}
	}
	return false
}
