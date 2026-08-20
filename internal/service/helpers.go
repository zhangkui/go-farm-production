package service

// defaultIfZero returns def when v is the zero value, else v.
func defaultIfZero[T comparable](v, def T) T {
	var zero T
	if v == zero {
		return def
	}
	return v
}
