package server

type mode int

const (
	autoMode mode = iota
	darkMode
	lightMode
)

func (m mode) String() string {
	return [...]string{"auto", "dark", "light"}[m]
}

func (param *Param) getMode() mode {
	if param.ForceDarkMode {
		return darkMode
	} else if param.ForceLightMode {
		return lightMode
	}
	return autoMode
}
