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
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	callout "gitlab.com/staticnoise/goldmark-callout"
	"go.abhg.dev/goldmark/anchor"
)

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
			&anchor.Extender{Texter: anchor.Text("#")},
			extension.GFM,
			emoji.Emoji,
			callout.CalloutExtention,
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
