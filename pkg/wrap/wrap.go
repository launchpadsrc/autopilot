package wrap

type Wrapped[T any] struct {
	v T
}

func Wrap[T any](v T) any {
	return Wrapped[T]{v: v}
}

func (w Wrapped[T]) Unwrap() T {
	return w.v
}

func Unwrap[T any](wrapped any) (zero T, _ bool) {
	w, ok := wrapped.(Wrapped[T])
	if !ok {
		return zero, false
	}
	return w.Unwrap(), true
}
