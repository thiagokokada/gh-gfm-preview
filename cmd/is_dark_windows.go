package cmd

import (
	"golang.org/x/sys/windows/registry"
)

func isDarkMode() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`, registry.QUERY_VALUE)
	if err != nil {
		logDebug("Debug [registry open key error]: %v", err)
		return isDarkModeDefault
	}
	defer k.Close()

	v, _, err := k.GetIntegerValue("AppsUseLightTheme")
	if err != nil {
		logDebug("Debug [get integer value error]: %v", err)
		return isDarkModeDefault
	}

	return v == 0 // 0 means dark mode, 1 means light mode
}
