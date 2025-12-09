package app

import (
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/thiagokokada/gh-gfm-preview/internal/assert"
)

func TestTargetFile(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"../../testdata/markdown-demo.md", "../../testdata/markdown-demo.md"},
		{"../../testdata/subdir/README.md", "../../testdata/subdir/README.md"},
		{"../../testdata/subdir", "../../testdata/subdir/README.md"},
	}
	for _, tt := range tests {
		actual, err := TargetFile(tt.input)
		assert.Nil(t, err)

		expected := tt.expected
		assert.Equal(t, actual, expected)
	}

	_, err := TargetFile("../../notfound.md")
	assert.NotNil(t, err)

	_, err = TargetFile("./")
	assert.NotNil(t, err)
}

func TestFindReadme(t *testing.T) {
	actual, err := FindReadme("../../testdata/subdir")
	assert.Nil(t, err)

	expected := "../../testdata/subdir/README.md"

	assert.Equal(t, actual, expected)

	actual, _ = FindReadme("../../testdata")
	expected = "../../testdata/README"

	assert.Equal(t, actual, expected)

	_, err = FindReadme("../../cmd")
	assert.NotNil(t, err)
}

func TestSlurp(t *testing.T) {
	result, err := Slurp("../../testdata/markdown-demo.md")
	assert.Nil(t, err)

	match := "Headings"
	r := regexp.MustCompile(match)

	assert.True(t, r.MatchString(result))

	_, err = Slurp("non-existing-file.md")
	assert.True(t, errors.Is(err, ErrFileNotFound))
}

func TestToHTML(t *testing.T) {
	markdown := "text"

	html, err := ToHTML(markdown, false)
	assert.Nil(t, err)

	actual := strings.TrimSpace(html)
	expected := "<p>text</p>"

	assert.Equal(t, actual, expected)
}

func TestGfmCheckboxes(t *testing.T) {
	result, err := Slurp("../../testdata/gfm-checkboxes.md")
	assert.Nil(t, err)

	html, err := ToHTML(result, false)
	assert.Nil(t, err)

	actual := strings.TrimSpace(html)

	checkBoxes := 0
	checkedCheckBoxes := 0
	uncheckedCheckBoxes := 0

	for line := range strings.SplitSeq(actual, "\n") {
		if strings.Contains(line, "type=\"checkbox\"") {
			checkBoxes++

			if strings.Contains(line, "checked") {
				checkedCheckBoxes++
			} else {
				uncheckedCheckBoxes++
			}
		}
	}

	assert.Equal(t, checkBoxes, 2)
	assert.Equal(t, checkedCheckBoxes, 1)
	assert.Equal(t, uncheckedCheckBoxes, 1)
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
		assert.True(t, strings.Contains(actual, target))
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
		assert.True(t, strings.Contains(actual, target))
	}
}

func TestParseExtensions(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"Empty string", "", []string{".md"}},
		{"Single extension with dot", ".txt", []string{".txt"}},
		{"Single extension without dot", "txt", []string{".txt"}},
		{"Multiple extensions", ".md,.txt,.rst", []string{".md", ".txt", ".rst"}},
		{"Multiple extensions without dots", "md,txt,rst", []string{".md", ".txt", ".rst"}},
		{"Wildcard", "*", []string{"*"}},
		{"Wildcard with spaces", " * ", []string{"*"}},
		{"Wildcard in middle", ".txt,*,.md", []string{"*"}},
		{"Wildcard at end", ".txt,.md,*", []string{"*"}},
		{"Mixed case", ".MD,.TxT", []string{".md", ".txt"}},
		{"With spaces", " .md , .txt ", []string{".md", ".txt"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseExtensions(tt.input)
			assert.Equal(t, len(got), len(tt.want))

			for i := range got {
				assert.Equal(t, got[i], tt.want[i])
			}
		})
	}
}
