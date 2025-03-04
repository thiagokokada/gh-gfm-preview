package server

import (
	"golang.org/x/sys/windows/registry"

	"github.com/thiagokokada/gh-gfm-preview/internal/utils"
)

func autoDetectDarkMode() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`, registry.QUERY_VALUE)
	if err != nil {
		utils.LogDebug("Debug [registry open key error]: %v", err)
		return isDarkModeDefault
	}
	defer k.Close()

	v, _, err := k.GetIntegerValue("AppsUseLightTheme")
	if err != nil {
		utils.LogDebug("Debug [get integer value error]: %v", err)
		return isDarkModeDefault
	}

	return v == 0 // 0 means dark mode, 1 means light mode
}
