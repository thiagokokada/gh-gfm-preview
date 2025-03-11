package browser

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type FileReader interface {
	ReadFile(filename string) (string, error)
}

type ProcVersionReader struct{}

func (r ProcVersionReader) ReadFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func isContainWSL(data string) bool {
	return strings.Contains(data, "WSL")
}

func isWSLWithReader(reader FileReader) bool {
	data, err := reader.ReadFile("/proc/version")
	if err != nil {
		return false
	}

	return isContainWSL(data)
}

func isWSL() bool {
	return isWSLWithReader(ProcVersionReader{})
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

	return exec.Command(cmd, args...).Start()
}
