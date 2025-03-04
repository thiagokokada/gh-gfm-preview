package cmd

import (
	"bytes"
	"os/exec"
	"strings"
)

func isDarkMode() bool {
	cmd := exec.Command("defaults", "read", "-g", "AppleInterfaceStyle")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return err == nil && strings.TrimSpace(out.String()) == "Dark"
}
