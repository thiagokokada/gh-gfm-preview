package browser

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type fileReader interface {
	readFile(filename string) (string, error)
}

type procVersionReader struct{}

func (r procVersionReader) readFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("OS read file error: %w", err)
	}

	return string(data), nil
}

func isContainWSL(data string) bool {
	return strings.Contains(data, "WSL")
}

func isWSLWithReader(reader fileReader) bool {
	data, err := reader.readFile("/proc/version")
	if err != nil {
		return false
	}

	return isContainWSL(data)
}

func isWSL() bool {
	return isWSLWithReader(procVersionReader{})
}

func OpenBrowser(url string) error {
	var args []string

	var cmd string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		if isWSL() {
			cmd = "cmd.exe"
			args = []string{"/c", "start"}
		} else {
			cmd = "xdg-open"
		}
	}

	args = append(args, url)

	err := exec.Command(cmd, args...).Start()
	if err != nil {
		return fmt.Errorf("exec command error: %w", err)
	}

	return nil
}
