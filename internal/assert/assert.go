package assert

import (
	"reflect"
	"testing"
)

func True(tb testing.TB, got bool) {
	tb.Helper()

	if !got {
		tb.Errorf("got: false; want: true")
	}
}

func False(tb testing.TB, got bool) {
	tb.Helper()

	if got {
		tb.Errorf("got: true; want: false")
	}
}

func Equal[T comparable](tb testing.TB, got, want T) {
	tb.Helper()

	if got != want {
		tb.Errorf("got: %v; want: %v", got, want)
	}
}

func Nil(tb testing.TB, got any) {
	tb.Helper()

	if !isNil(got) {
		tb.Errorf("got: %v; want: nil", got)
	}
}

func NotNil(tb testing.TB, got any) {
	tb.Helper()

	if isNil(got) {
		tb.Errorf("got: %v; want: !nil", got)
	}
}

func isNil(v any) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() { //nolint:exhaustive
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	}

	return false
}
