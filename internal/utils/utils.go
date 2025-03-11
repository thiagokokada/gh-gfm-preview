package utils

import "log"

var Verbose bool

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}

func LogInfo(format string, v ...any) {
	log.Printf(format, v...)
}

func LogDebug(format string, v ...any) {
	if Verbose {
		log.Printf(format, v...)
	}
}
