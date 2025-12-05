package utils

import "log"

var verbose bool

func Must[T any](v T, err error) T { //nolint:ireturn
	if err != nil {
		panic(err)
	}

	return v
}

func LogInfof(format string, v ...any) {
	log.Printf(format, v...)
}

func LogDebugf(format string, v ...any) {
	if verbose {
		log.Printf("DEBUG "+format, v...)
	}
}

func SetVerbose(v bool) {
	verbose = v
}
