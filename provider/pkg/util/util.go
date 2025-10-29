package util

func OrZero[T any](v *T) (r T) {
	if v == nil {
		return r
	}
	return *v
}
