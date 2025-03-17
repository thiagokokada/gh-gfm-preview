package server

import (
	"github.com/thiagokokada/dark-mode-go"
	"github.com/thiagokokada/gh-gfm-preview/internal/utils"
)

const (
	darkMode    = "dark"
	lightMode   = "light"
	defaultMode = darkMode
)

func (param *Param) getMode() string {
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
