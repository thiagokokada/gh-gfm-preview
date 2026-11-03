# Print layout

Preview this file with `go run . --light-mode testdata/print.md`, then open
the browser's print preview or save it as a PDF. Check portrait A4 and Letter
paper at 100% scale. Repeat in directory mode with
`go run . --light-mode --directory-listing testdata`, then open `print.md`.

All table columns and the ends of long code lines should be visible. Copy,
diagram, and navigation controls should be hidden on paper. On screen, wide
content should still scroll horizontally and controls should remain available.

## Wide table

The rightmost column must be visible without reducing the print scale.

| Component | Purpose | Input | Output | Configuration | Owner | Status | Rightmost column |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Markdown renderer | Convert source documents into formatted HTML | Markdown with tables and code | HTML for the browser | GitHub Flavored Markdown | Rendering team | Ready for printing | TABLE-END-ONE |
| Directory browser | Navigate documents in nested directories | A local directory tree | File links and breadcrumbs | Custom file extensions | Navigation team | Ready for printing | TABLE-END-TWO |
| Live reload | Refresh previews after source changes | Filesystem notifications | Updated document content | Automatic file watching | Runtime team | Ready for printing | TABLE-END-THREE |

## Long cell contents

Long URLs and inline code should wrap within their cells.

| Kind | Value | End marker |
| --- | --- | --- |
| URL | https://example.com/documentation/very-long-path-without-spaces/abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789 | URL-END |
| Inline code | `abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789` | INLINE-END |

## Plain code block

The following line should wrap on paper, including its final marker.

```text
This deliberately long line exercises printing without horizontal clipping: alpha beta gamma delta epsilon zeta eta theta iota kappa lambda mu nu xi omicron pi rho sigma tau upsilon phi chi psi omega PLAIN-CODE-END
abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789 TOKEN-END
```

## Highlighted code block

Syntax highlighting should preserve indentation while allowing long lines to
wrap on paper.

```go
package main

import "fmt"

func main() {
    fmt.Println("This deliberately long highlighted string should wrap within the printable page instead of losing its final words: alpha beta gamma delta epsilon zeta eta theta iota kappa lambda HIGHLIGHTED-CODE-END")
}
```

## Image and page breaks

The image should stay together on one page, and headings should stay with the
content that follows them when space allows.

![Dinotocat print sample](images/dinotocat.png)

### Final heading

This paragraph should stay with its heading. PRINT-FIXTURE-END
