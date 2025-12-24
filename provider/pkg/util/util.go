package util

import (
	"cmp"
	"slices"
)

func OrZero[T any](v *T) (r T) {
	if v == nil {
		return r
	}
	return *v
}

// OrNil returns nil if v is the zero value for its type, otherwise returns a pointer to v.
func OrNil[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}
	return &v
}

func ElementsEqual[T cmp.Ordered](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	aCopy := make([]T, len(a))
	bCopy := make([]T, len(b))
	copy(aCopy, a)
	copy(bCopy, b)
	slices.Sort(aCopy)
	slices.Sort(bCopy)
	return slices.Equal(aCopy, bCopy)
}

// ElementsEqualFunc compares two slices for equality, ignoring order.
// It uses the provided cmp function for sorting and eq function for equality comparison.
// The cmp function should return a negative number when a < b, zero when a == b,
// and a positive number when a > b.
func ElementsEqualFunc[T any](a, b []T, cmpFunc func(a, b T) int, eq func(a, b T) bool) bool {
	if len(a) != len(b) {
		return false
	}
	aCopy := make([]T, len(a))
	bCopy := make([]T, len(b))
	copy(aCopy, a)
	copy(bCopy, b)
	slices.SortFunc(aCopy, cmpFunc)
	slices.SortFunc(bCopy, cmpFunc)
	return slices.EqualFunc(aCopy, bCopy, eq)
}
