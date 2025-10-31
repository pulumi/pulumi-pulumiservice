package util

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
