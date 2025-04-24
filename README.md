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

## Screenshots

Open your browser:

<img width="2013" alt="Screenshot 2025-04-24 at 19 44 45" src="https://github.com/user-attachments/assets/d31d3665-9f75-4ddc-a7c9-526c561eb472" />

Live reloading:

https://github.com/user-attachments/assets/0219ac01-71d3-4568-bff4-c9de092ca4e3

## Highlights

- **Works offline** - You don't need an internet connection.
- **Fast** - Since it doesn't rely on external services it is really fast.
- **No-dependencies** - You can just run the standalone binary (or optionally
  via `gh` as an extension).
- **Zero-configuration** - You don't have to set the GitHub access token.
- **Live reloading** - You don't need reload the browser.
- **Auto open browser** - Your browser will be opened automatically.
- **Auto find port** - You don't need find an available port if default is used.
- **Graceful degradation** - Basic functionality works even without JavaScript.

## Supported GFM features

- [x] : [Most (all?) of GitHub Flavored Markdown spec](https://github.github.com/gfm/)
- [x] : [Emojis](https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax#using-emojis)
  + not all of GitHub emojis are supported since a few of them are extensions
    to the Unicode spec, but the majority of them works
- [x] : [Alerts](https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax#alerts)
- [x] : [Code blocks with syntax
  highlighting](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/creating-and-highlighting-code-blocks)
  + highlighting is done via
    [alecthomas/chroma](https://github.com/alecthomas/chroma); not all GitHub
    languages are supported and there are slightly differences in highlighting
- [x] : [Section links](https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax#section-links)
- [x] : [Raw HTML](https://gist.github.com/seanh/13a93686bf4c2cb16e658b3cf96807f2)
  + no filtering is done so you can pass any arbitrary HTML, different from
    GitHub where only a subset of HTML is allowed
- [x] : [MathJax](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/writing-mathematical-expressions)
- [x] : [Mermaid diagrams](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/creating-diagrams)
- [ ] : [GeoJSON/TopoJSON diagrams](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/creating-diagrams)
- [ ] : [STL 3D diagrams](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/creating-diagrams)
- [ ] : Mentioning people/Referencing issues and Pull Requests/Other features
  that depends in GitHub access

## Installation

### GitHub Extension

You need to have [gh](https://github.com/cli/cli#installation) installed.

```console
gh extension install thiagokokada/gh-gfm-preview
```

Upgrade:

```console
gh extension upgrade markdown-preview
```

### Standalone

You need to have [Go](https://go.dev/) installed.

```console
go install github.com/thiagokokada/gh-gfm-preview@latest
```

### Nix

Assuming that you have [Flakes](https://wiki.nixos.org/wiki/Flakes) enabled:

```console
nix run github:thiagokokada/gh-gfm-preview
```

## Usage

The usage:

```console
gh gfm-preview README.md
```

Or this command will detect README file in the directory automatically.

```console
gh gfm-preview
```

Then access the local web server such as `http://localhost:3333` with Chrome,
Firefox, or Safari.

Available options:

```
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

## Other usages

Since the binary is static and it works offline, this is a good program to
use to preview how a Markdown is looking in e.g.:
[neovim](https://github.com/neovim/neovim/). For example, you can add this
in your `$HOME/.config/nvim/init.lua`:

```lua
local function preview_markdown()
  local file = vim.fn.expand("%")
  local on_exit_cb = function(out)
    print("Markdown preview process exited with code:", out.code)
  end
  local process = vim.system(
    -- assuming that the extension were installed using gh
    -- the reason we are not using `gh gfm-preview` instead is because this
    -- can cause an issue where the gh process is killed but not the
    -- gh-gfm-preview, since the kill signal will not reach the child process
    {vim.fn.expand("$HOME/.local/share/gh/extensions/gh-gfm-preview/gh-gfm-preview"), file},
    on_exit_cb
  )

  vim.api.nvim_create_autocmd({ "BufUnload", "BufDelete" }, {
    buffer = vim.api.nvim_get_current_buf(),
    callback = function()
      process:kill("sigterm")
      -- timeout (in ms), will call SIGKILL upon timeout
      process:wait(500)
    end,
  })
end

-- create a shortcut only in Markdown files, mapped to `<Leader>P`
vim.api.nvim_create_autocmd({ "FileType" }, {
  pattern = { "markdown" },
  callback = function()
    vim.keymap.set("n", "<Leader>P", preview_markdown, {
      desc = "Markdown preview", buffer = true
    })
  end,
})
```

## Development

You can run the following command to (re-)generate assets:

```console
go generate ./...
```

And you can run the following command to build:

```console
go build
```

## Related projects

- GitHub CLI <https://cli.github.com>
- Grip <https://github.com/joeyespo/grip>
- github-markdown-css <https://github.com/sindresorhus/github-markdown-css>
- gh-markdown-preview <https://github.com/yusukebe/gh-markdown-preview/>
