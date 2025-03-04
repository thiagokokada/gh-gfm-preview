//go:build !windows && !darwin && !linux
package server

func autoDetectDarkMode() bool {
	return isDarkModeDefault
}
