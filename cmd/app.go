package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	"github.com/yuin/goldmark/extension"
	"gitlab.com/staticnoise/goldmark-callout"
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
	ext := goldmark.WithExtensions()
	if !param.markdownMode {
		ext = goldmark.WithExtensions(extension.GFM, emoji.Emoji, callout.CalloutExtention)
	}
	md := goldmark.New(ext)
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
