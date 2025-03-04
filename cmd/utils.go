package cmd

import "log"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func must2[T any](v T, err error) T {
	must(err)
	return v
}

func logInfo(format string, v ...any) {
	log.Printf(format, v...)
}

func logDebug(format string, v ...any) {
	if verbose {
		log.Printf(format, v...)
	}
}
