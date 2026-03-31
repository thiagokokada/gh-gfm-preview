package app

import (
	"strconv"

	"github.com/yuin/goldmark"
	ast "github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

const footnoteBacklinkHTML = "&#x21a9;&#xfe0e;"

type footnoteBacklinkExtender struct{}

func newFootnoteExtender() *footnoteBacklinkExtender {
	return &footnoteBacklinkExtender{}
}

func (e *footnoteBacklinkExtender) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&footnoteBacklinkHTMLRenderer{}, 400),
	))
}

type footnoteBacklinkHTMLRenderer struct{}

func (r *footnoteBacklinkHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(extast.KindFootnoteBacklink, r.renderFootnoteBacklink)
}

func (r *footnoteBacklinkHTMLRenderer) renderFootnoteBacklink(
	w util.BufWriter, _ []byte, node ast.Node, entering bool,
) (ast.WalkStatus, error) {
	if entering {
		n, ok := node.(*extast.FootnoteBacklink)
		if !ok {
			return ast.WalkStop, nil
		}

		is := strconv.Itoa(n.Index)

		_, _ = w.WriteString(`&#160;<a href="#fnref`)
		if n.RefIndex > 0 {
			_, _ = w.WriteString(strconv.Itoa(n.RefIndex))
		}

		_, _ = w.WriteString(`:`)
		_, _ = w.WriteString(is)
		_, _ = w.WriteString(`" role="doc-backlink">`)

		_, _ = w.WriteString(footnoteBacklinkHTML)

		if n.RefIndex > 0 {
			_, _ = w.WriteString(`<sup>`)
			_, _ = w.WriteString(strconv.Itoa(n.RefIndex + 1))
			_, _ = w.WriteString(`</sup>`)
		}

		_, _ = w.WriteString(`</a>`)
	}

	return ast.WalkContinue, nil
}
