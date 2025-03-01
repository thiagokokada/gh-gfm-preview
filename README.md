# gh gfm-preview

A Go program to preview GitHub Flavored Markdown (GFM) :notebook:.

The `gh-gfm-preview` command start a local web server to serve the markdown
document. **gh gfm-preview** renders the HTML using
[yuin/goldmark](https://github.com/yuin/goldmark) and some extensions and
frontend tricks to have similar features and look to how GitHub renders a
markdown.

It may also be used as a [GitHub CLI](https://cli.github.com) extension.

This is a hard fork of
[yusukebe/gh-markdown-preview](https://github.com/yusukebe/gh-markdown-preview/),
that uses the [GitHub Markdown API](https://docs.github.com/en/rest/markdown),
but this means it doesn't work offline. The code of this repository tries to
emulate the look of GitHub Markdown rendering as close as possible, but the
original project will be even closer to the actual result if you don't need
offline rendering.

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
go run github.com/thiagokokada/gh-gfm-preview
```

### Nix

Assuming that you have [Flakes](https://wiki.nixos.org/wiki/Flakes) enabled:

```
nix run github:thiagokokada/gh-gfm-preview
```

### GitHub Extension

```
gh extension install thiagokokada/gh-gfm-preview
```

Upgrade:

```
gh extension upgrade markdown-preview
```

## Usage

The usage:

```
gh gfm-preview README.md
```

Or this command will detect README file in the directory automatically.

```
gh gfm-preview
```

Then access the local web server such as `http://localhost:3333` with Chrome,
Firefox, or Safari.

Available options:

```text
    --dark-mode           force dark mode
    --disable-auto-open   disable auto opening your browser
    --disable-reload      disable live reloading
-h, --help                help for gh-gfm-preview
    --host string         hostname this server will bind (default "localhost")
    --light-mode          force light mode
    --markdown-mode       force "markdown" mode (rather than default "gfm")
-p, --port int            TCP port number of this server (default 3333)
    --verbose             show verbose output
```

## Related projects

- GitHub CLI <https://cli.github.com>
- Grip <https://github.com/joeyespo/grip>
- github-markdown-css <https://github.com/sindresorhus/github-markdown-css>
- gh-markdown-preview <https://github.com/yusukebe/gh-markdown-preview/>
