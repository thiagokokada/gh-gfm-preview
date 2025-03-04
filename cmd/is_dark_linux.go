package cmd

import (
	"os/exec"
	"strings"
)

func isDarkMode() bool {
	cmd := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "color-scheme")
	out, err := cmd.Output()
	if err != nil {
		return isDarkModeDefault
	}
	return strings.Contains(string(out), "dark")
}
