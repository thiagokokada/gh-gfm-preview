package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/styles"
	// TODO: switch to upstream once this is merged: https://github.com/abhinav/goldmark-anchor/pull/74
	anchor "github.com/thiagokokada/goldmark-anchor"
	alerts "github.com/thiagokokada/goldmark-gh-alerts"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

const anchorIcon = `<svg class="octicon octicon-link" viewBox="0 0 16 16" version="1.1" width="16" height="16" aria-hidden="true"><path d="m7.775 3.275 1.25-1.25a3.5 3.5 0 1 1 4.95 4.95l-2.5 2.5a3.5 3.5 0 0 1-4.95 0 .751.751 0 0 1 .018-1.042.751.751 0 0 1 1.042-.018 1.998 1.998 0 0 0 2.83 0l2.5-2.5a2.002 2.002 0 0 0-2.83-2.83l-1.25 1.25a.751.751 0 0 1-1.042-.018.751.751 0 0 1-.018-1.042Zm-4.69 9.64a1.998 1.998 0 0 0 2.83 0l1.25-1.25a.751.751 0 0 1 1.042.018.751.751 0 0 1 .018 1.042l-1.25 1.25a3.5 3.5 0 1 1-4.95-4.95l2.5-2.5a3.5 3.5 0 0 1 4.95 0 .751.751 0 0 1-.018 1.042.751.751 0 0 1-1.042.018 1.998 1.998 0 0 0-2.83 0l-2.5 2.5a1.998 1.998 0 0 0 0 2.83Z"></path></svg>`

var github = must2(
	styles.Get("github").Builder().AddEntry(
		chroma.Background, chroma.StyleEntry{
			Background: chroma.NewColour(246, 248, 250),
		},
	).Build(),
)
var githubDark = must2(
	styles.Get("github-dark").Builder().AddEntry(
		chroma.Background, chroma.StyleEntry{
			Background: chroma.NewColour(21, 27, 35),
		},
	).Build(),
)

func targetFile(filename string) (string, error) {
	var err error
	if filename == "" {
		filename = "."
	}
	info, err := os.Stat(filename)
	if err == nil && info.IsDir() {
		readme, err := findReadme(filename)
		if err != nil {
			return "", err
		}
		filename = readme
	}
	if err != nil {
		err = fmt.Errorf("%s is not found", filename)
	}
	return filename, err
}

func findReadme(dir string) (string, error) {
	files, _ := os.ReadDir(dir)
	for _, f := range files {
		r := regexp.MustCompile(`(?i)^readme`)
		if r.MatchString(f.Name()) {
			return filepath.Join(dir, f.Name()), nil
		}
	}
	err := fmt.Errorf("README file is not found in %s/", dir)
	return "", err
}

func toHTML(markdown string, param *Param) (string, error) {
	style := githubDark
	if getMode(param) == lightMode {
		style = github
	}
	ext := goldmark.WithExtensions()
	if !param.markdownMode {
		ext = goldmark.WithExtensions(
			&anchor.Extender{Texter: anchor.Text(anchorIcon)},
			alerts.GhAlertsExtension,
			extension.GFM,
			emoji.Emoji,
			highlighting.NewHighlighting(highlighting.WithCustomStyle(style)),
		)
	}
	md := goldmark.New(ext, goldmark.WithParserOptions(parser.WithAutoHeadingID()))
	var buf bytes.Buffer
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func slurp(fileName string) (string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer f.Close()
	b, _ := io.ReadAll(f)
	text := string(b)
	return text, nil
}
