package app

import (
	"errors"
	"regexp"
	"strings"
	"testing"
)

func TestTargetFile(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"../../testdata/markdown-demo.md", "../../testdata/markdown-demo.md"},
		{"../../README.md", "../../README.md"},
		{"../../", "../../README.md"},
	}
	for _, tt := range tests {
		actual, err := TargetFile(tt.input)
		if err != nil {
			t.Errorf("%s", err.Error())
		}

		expected := tt.expected
		if actual != expected {
			t.Errorf("got %v\n want %v", actual, expected)
		}
	}

	_, err := TargetFile("../../notfound.md")
	if err == nil {
		t.Errorf("err is nil")
	}

	_, err = TargetFile("./")
	if err == nil {
		t.Errorf("err is nil")
	}
}

func TestFindReadme(t *testing.T) {
	actual, _ := findReadme("../../")
	expected := "../../README.md"

	if actual != expected {
		t.Errorf("got %v\n want %v", actual, expected)
	}

	actual, _ = findReadme("../../testdata")
	expected = "../../testdata/README"

	if actual != expected {
		t.Errorf("got %v\n want %v", actual, expected)
	}

	_, err := findReadme("../../cmd")
	if err == nil {
		t.Errorf("err is nil")
	}
}

func TestSlurp(t *testing.T) {
	result, err := Slurp("../../testdata/markdown-demo.md")
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	match := "Headings"
	r := regexp.MustCompile(match)

	if r.MatchString(result) == false {
		t.Errorf("content do not match %v\n", match)
	}

	_, err = Slurp("non-existing-file.md")
	if !errors.Is(err, ErrFileNotFound) {
		t.Errorf("wrong error for non-existing-file %v\n", err)
	}
}

func TestToHTML(t *testing.T) {
	markdown := "text"

	html, err := ToHTML(markdown, false)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	actual := strings.TrimSpace(html)
	expected := "<p>text</p>"

	if actual != expected {
		t.Errorf("got %v\n want %v", actual, expected)
	}
}

func TestGfmCheckboxes(t *testing.T) {
	result, err := Slurp("../../testdata/gfm-checkboxes.md")
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	html, err := ToHTML(result, false)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	actual := strings.TrimSpace(html)

	checkBoxes := 0
	checkedCheckBoxes := 0
	uncheckedCheckBoxes := 0

	for _, line := range strings.Split(actual, "\n") {
		if strings.Contains(line, "type=\"checkbox\"") {
			checkBoxes++

			if strings.Contains(line, "checked") {
				checkedCheckBoxes++
			} else {
				uncheckedCheckBoxes++
			}
		}
	}

	if checkBoxes != 2 {
		t.Errorf("got %v checkboxes, want 2", checkBoxes)
	}

	if checkedCheckBoxes != 1 {
		t.Errorf("got %v checked checkboxes, want 1", checkedCheckBoxes)
	}

	if uncheckedCheckBoxes != 1 {
		t.Errorf("got %v unchecked checkboxes, want 1", uncheckedCheckBoxes)
	}
}

func TestGfmAlerts(t *testing.T) {
	result, err := Slurp("../../testdata/gfm-alerts.md")
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	html, err := ToHTML(result, false)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	actual := strings.TrimSpace(html)

	for _, target := range []string{
		"markdown-alert-note",
		"markdown-alert-tip",
		"markdown-alert-important",
		"markdown-alert-warning",
		"markdown-alert-caution",
	} {
		if !strings.Contains(actual, target) {
			t.Errorf("expected but not found: %s", target)
		}
	}
}

func TestRawHTML(t *testing.T) {
	result, err := Slurp("../../testdata/markdown-demo.md")
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	html, err := ToHTML(result, false)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	actual := strings.TrimSpace(html)

	for _, target := range []string{
		`<p align="center">`,
		"<details>",
		"<summary>",
		`<sup id="backToMyFootnote">`,
	} {
		if !strings.Contains(actual, target) {
			t.Errorf("expected but not found: %s", target)
		}
	}
}
