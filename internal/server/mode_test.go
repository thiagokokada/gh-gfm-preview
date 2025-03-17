package server

import "testing"

func TestGetMode(t *testing.T) {
	param := &Param{ForceLightMode: true}
	modeString := param.getMode().String()
	expected := "light"

	if modeString != expected {
		t.Errorf("mode string is not: %s", modeString)
	}

	param = &Param{ForceDarkMode: true}
	modeString = param.getMode().String()
	expected = "dark"

	if modeString != expected {
		t.Errorf("mode string is not: %s", modeString)
	}
}
