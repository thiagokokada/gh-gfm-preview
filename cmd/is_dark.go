//go:build !windows && !darwin && !linux
package cmd

func isDarkMode() bool {
	return isDarkModeDefault
}
