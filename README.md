# GFM markdown-preview

A Go program to preview GitHub Flavored Markdown :octocat:.

The `gfm-markdown-preview` command start a local web server to serve the
markdown document. **gfm-markdown-preview** renders the HTML using
[yuin/goldmark](https://github.com/yuin/goldmark) and some extensions to have
similar features and look to how GitHub renders a markdown.

It may also be used as a [GitHub CLI](https://cli.github.com) extension.

This is a hard fork of
[yusukebe/gh-markdown-preview](https://github.com/yusukebe/gh-markdown-preview/),
that uses the [GitHub Markdown API](https://docs.github.com/en/rest/markdown),
but this means it doesn't work offline.

## Features

- **Works offline** - You don't need an internet connection.
- **No-dependencies** - You need `gh` command only.
- **Zero-configuration** - You don't have to set the GitHub access token.
- **Live reloading** - You don't need reload the browser.
- **Auto open browser** - Your browser will be opened automatically.
- **Auto find port** - You don't need find an available port if default is used.

## Installation

You need to have [Go](https://go.dev/) installed.

### Standalone

```
go run github.com/thiagokokada/gfm-markdown-preview
```

### GitHub Extension

```
gh extension install thiagokokada/gfm-markdown-preview
```

Upgrade:

```
gh extension upgrade markdown-preview
```

## Usage

The usage:

```
gfm-markdown-preview README.md
```

Or this command will detect README file in the directory automatically.

```
gfm-markdown-preview
```

Then access the local web server such as `http://localhost:3333` with Chrome,
Firefox, or Safari.

Available options:

```text
    --dark-mode           Force dark mode
    --markdown-mode       Force "markdown" mode (rather than default "gfm")
    --disable-auto-open   Disable auto opening your browser
    --disable-reload      Disable live reloading
-h, --help                help for gfm-markdown-preview
    --host string         Hostname this server will bind (default "localhost")
    --light-mode          Force light mode
-p, --port int            TCP port number of this server (default 3333)
    --verbose             Show verbose output
    --version             Show the version
```

## Related projects

- GitHub CLI <https://cli.github.com>
- Grip <https://github.com/joeyespo/grip>
- github-markdown-css <https://github.com/sindresorhus/github-markdown-css>
- gh-markdown-preview <https://github.com/yusukebe/gh-markdown-preview/>
