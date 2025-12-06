package utils

func Must[T any](v T, err error) T { //nolint:ireturn
	if err != nil {
		panic(err)
	}

	return v
}
