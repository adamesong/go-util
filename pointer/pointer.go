package pointer

func New[T any](value T) *T {
	return &value
}
func Value[T any](ptr *T) T {
	if ptr == nil {
		var zero T
		return zero
	}
	return *ptr
}
func ValueOr[T any](ptr *T, def T) T {
	if ptr == nil {
		return def
	}
	return *ptr
}
func ValueOrZero[T any](ptr *T) T {
	return ValueOr(ptr, *new(T))
}
func IsNil[T any](ptr *T) bool {
	return ptr == nil
}
func IsNotNil[T any](ptr *T) bool {
	return ptr != nil
}
func IsZero[T comparable](ptr *T) bool {
	return ptr == nil || *ptr == *new(T)
}
func IsNotZero[T comparable](ptr *T) bool {
	return ptr != nil && *ptr != *new(T)
}
func IsEqual[T comparable](ptr *T, value T) bool {
	return ptr != nil && *ptr == value
}
func IsNotEqual[T comparable](ptr *T, value T) bool {
	return ptr == nil || *ptr != value
}
func IsEqualZero[T comparable](ptr *T) bool {
	return ptr == nil || *ptr == *new(T)
}
func IsNotEqualZero[T comparable](ptr *T) bool {
	return ptr != nil && *ptr != *new(T)
}
func IsEqualOrZero[T comparable](ptr *T, value T) bool {
	return ptr == nil || *ptr == value || *ptr == *new(T)
}
func IsNotEqualOrZero[T comparable](ptr *T, value T) bool {
	return ptr != nil && *ptr != value && *ptr != *new(T)
}

func Set[T any](ptr *T, value T) {
	if ptr != nil {
		*ptr = value
	}
}
func SetZero[T any](ptr *T) {
	if ptr != nil {
		*ptr = *new(T)
	}
}

func SetDefault[T any](ptr *T, def T) {
	if ptr != nil {
		*ptr = def
	}
}
func SetDefaultZero[T any](ptr *T) {
	if ptr != nil {
		*ptr = *new(T)
	}
}
