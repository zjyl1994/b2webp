package utils

func Map[T, R any](inputs []T, fn func(T) R) []R {
	result := make([]R, len(inputs))
	for i, v := range inputs {
		result[i] = fn(v)
	}
	return result
}

func Convert[T, R any](inputs []T, fn func(T) (R, error)) ([]R, error) {
	result := make([]R, len(inputs))
	for i, v := range inputs {
		val, err := fn(v)
		if err != nil {
			return nil, err
		}
		result[i] = val
	}
	return result, nil
}

func Contains[T comparable](item T, list []T) bool {
	for _, v := range list {
		if item == v {
			return true
		}
	}
	return false
}

func Coalesce[T comparable](s ...T) T {
	var empty T
	for _, v := range s {
		if v != empty {
			return v
		}
	}
	return empty
}
