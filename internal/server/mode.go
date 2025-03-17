package server

import (
	"github.com/thiagokokada/dark-mode-go"
	"github.com/thiagokokada/gh-gfm-preview/internal/utils"
)

type mode int

const (
	darkMode mode = iota
	lightMode
	defaultMode = darkMode
)

func (m mode) String() string {
	return [...]string{"dark", "light"}[m]
}

func (param *Param) getMode() mode {
	if param.ForceDarkMode {
		return darkMode
	} else if param.ForceLightMode {
		return lightMode
	}

	isDark, err := dark.IsDarkMode()
	utils.LogDebug("Debug [auto-detected dark mode]: isDark=%v, err=%v", isDark, err)

	if err != nil {
		return defaultMode
	}

	if isDark {
		return darkMode
	}

	return lightMode
}

func (param *Param) isDarkMode() bool {
	return param.getMode() == darkMode
}
