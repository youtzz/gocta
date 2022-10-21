package internal

func Max[T int | float64](x, y T) T {
	if x > y {
		return x
	} else {
		return y
	}
}

func Min[T int | float64](x, y T) T {
	if x < y {
		return x
	} else {
		return y
	}
}
