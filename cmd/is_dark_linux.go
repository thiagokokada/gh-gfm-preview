package cmd

import (
	"github.com/godbus/dbus/v5"
)

func isDarkMode() bool {
	conn, err := dbus.SessionBus()
	if err != nil {
		logDebug("Debug [dbus connect session bus error]: %v", err)
		return isDarkModeDefault
	}
	defer conn.Close()

	obj := conn.Object("org.freedesktop.portal.Desktop", "/org/freedesktop/portal/desktop")
	var colorScheme uint32
	err = obj.Call("org.freedesktop.portal.Settings.Read", 0, "org.freedesktop.appearance", "color-scheme").Store(&colorScheme)
	if err != nil {
		logDebug("Debug [dbus call error]: %v", err)
		return isDarkModeDefault
	}
	return colorScheme == 0 || colorScheme == 1 // 0: no preference, 1: prefer dark mode, 2: prefer light mode
}
