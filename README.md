# gh gfm-preview

A Go program to preview GitHub Flavored Markdown (GFM) :notebook:.

The `gh-gfm-preview` command starts a local web server that serves the Markdown
document. **gh gfm-preview** renders HTML using
[yuin/goldmark](https://github.com/yuin/goldmark), extensions, and frontend
tricks to provide features and an appearance similar to GitHub's Markdown
rendering.

It may also be used as a [GitHub CLI](https://cli.github.com) extension.

This is a hard fork of
[yusukebe/gh-markdown-preview](https://github.com/yusukebe/gh-markdown-preview/),
which uses the [GitHub Markdown API](https://docs.github.com/en/rest/markdown)
and therefore does not work offline. This repository aims to emulate GitHub's
Markdown rendering as closely as possible, but the original project produces a
closer match when offline rendering is not required.

## Screenshots

Open your browser:

![Screenshot showing a Markdown document being edited on the left and rendered
with gh-gfm-preview on the right](screenshot.png)

Live reloading:

https://github.com/user-attachments/assets/0219ac01-71d3-4568-bff4-c9de092ca4e3

## Highlights

- **Works offline** - You don't need an internet connection to use most
  features.
- **Fast** - Since it doesn't rely on external services, it is very fast.
- **No dependencies** - You can just run the standalone binary (or optionally
  via `gh` as an extension).
- **Zero-configuration** - You don't have to set the GitHub access token.
- **Live reloading** - You don't need to reload the browser.
- **Auto open browser** - Your browser will be opened automatically.
- **Automatic port selection** - You don't need to find an available port when
  using the default.
- **Graceful degradation** - Basic functionality works even without JavaScript.

## Supported GFM features

| Feature | Supported | Available Offline | Needs JavaScript | Notes |
| --- | --- | --- | --- | --- |
| [GitHub Flavored Markdown specification](https://github.github.com/gfm/) | Yes | Yes | No | Most of the specification should be supported, thanks to [yuin/goldmark](https://github.com/yuin/goldmark) and its extensions. |
| [Emojis](https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax#using-emojis) | Yes | Yes | No | Not all GitHub emojis are supported because some are extensions to the Unicode specification, but most work. |
| [Alerts](https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax#alerts) | Yes | Yes | No | Custom alert labels such as `[!UNKNOWN]` are supported but render differently from GitHub. |
| [Code blocks with syntax highlighting](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/creating-and-highlighting-code-blocks) | Yes | Yes | No | Highlighting uses [alecthomas/chroma](https://github.com/alecthomas/chroma). Not all GitHub languages are supported, and there are slight differences in highlighting. |
| [Section links](https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax#section-links) | Yes | Yes | No | |
| [Raw HTML](https://gist.github.com/seanh/13a93686bf4c2cb16e658b3cf96807f2) | Yes | Yes | No | No filtering is performed, so arbitrary HTML is allowed. GitHub allows only a subset of HTML. |
| [MathJax](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/writing-mathematical-expressions) | Yes | Yes | Yes | |
| [Mermaid diagrams](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/creating-diagrams) | Yes | Yes | Yes | |
| [GeoJSON/TopoJSON diagrams](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/creating-diagrams) | Yes | No | Yes | Rendered with Leaflet and online map tiles. The basemap is OSM-style rather than GitHub’s Azure/TomTom tiles, so it is structurally similar rather than pixel-identical. |
| [STL 3D diagrams](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/creating-diagrams) | No | N/A | N/A | |
| Mentioning people, referencing issues and pull requests, and other features that depend on GitHub access | No | N/A | N/A | Out of scope because they require GitHub API access. |

## Installation

### GitHub Extension

You need to have [gh](https://github.com/cli/cli#installation) installed.

```console
gh extension install thiagokokada/gh-gfm-preview
```

Upgrade:

```console
gh extension upgrade gfm-preview
```

### Standalone

You need to have [Go](https://go.dev/) installed.

```console
go install github.com/thiagokokada/gh-gfm-preview@latest
```

### Nix

Assuming that you have [Nix flakes](https://wiki.nixos.org/wiki/Flakes)
enabled:

```console
nix run github:thiagokokada/gh-gfm-preview
```

## Usage

To preview a file:

```console
gh gfm-preview README.md
```

To automatically detect a README file in the current directory:

```console
gh gfm-preview
```

You can also preview Markdown from standard input by piping data or using `-`:

```
echo "# Hello" | gh gfm-preview
cat README.md | gh gfm-preview
gh gfm-preview - < README.md
```

Then open the local web server, for example at `http://localhost:3333`, in
your default browser.

Available options:

```
  -p, --port int                                   TCP port number of this server (default 3333)
  -H, --host string                                hostname this server will bind (default "localhost")
  -R, --disable-reload                             disable live reloading
  -A, --disable-auto-open                          disable auto opening your browser
  -l, --light-mode                                 force light mode
  -d, --dark-mode                                  force dark mode
  -m, --markdown-mode                              force "markdown" mode (rather than default "gfm")
  -D, --directory-listing                          enable directory browsing mode
      --directory-listing-show-extensions string   file extensions to show in directory listing (comma-separated, use '*' for all files) (default ".md,.txt")
      --directory-listing-text-extensions string   text file extensions for preview (comma-separated, others will be served as binary) (default ".md,.txt")
      --no-color                                   disable color for logs
  -v, --verbose                                    show verbose output
      --version                                    show program version
```

### Directory Listing

Enable directory browsing mode to navigate and preview files:

```console
# Enable with default settings
gh gfm-preview --directory-listing
# Or
gh gfm-preview -D

# Show all file types
gh gfm-preview --directory-listing --directory-listing-show-extensions="*"

# Custom file extensions
gh gfm-preview --directory-listing \
  --directory-listing-show-extensions=".md,.rst,.adoc" \
  --directory-listing-text-extensions=".md,.txt,.rst"
```

## Other usages

Because the binary is static and works offline, it is well suited to previewing
how Markdown will look in applications such as
[Neovim](https://github.com/neovim/neovim/). For example, you can add this
in your `$HOME/.config/nvim/init.lua`:

```lua
local function preview_markdown()
  local file = vim.fn.expand("%")
  local on_exit_cb = function(out)
    print("Markdown preview process exited with code:", out.code)
  end
  local process = vim.system(
    -- assuming that the extension was installed using gh
    -- the reason we are not using `gh gfm-preview` is that this can cause an
    -- issue where the gh process is killed but gh-gfm-preview is not, since
    -- the kill signal does not reach the child process
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

And you can run the following command to build the project:

```console
go build
```

If you have `nix` with [Flakes](https://wiki.nixos.org/wiki/Flakes) enabled:

```console
nix develop
```

## Related projects

- GitHub CLI <https://cli.github.com>
- Grip <https://github.com/joeyespo/grip>
- github-markdown-css <https://github.com/sindresorhus/github-markdown-css>
- gh-markdown-preview <https://github.com/yusukebe/gh-markdown-preview/>
