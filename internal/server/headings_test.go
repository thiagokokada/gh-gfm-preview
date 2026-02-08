package server

import (
	"strings"
	"testing"

	"github.com/thiagokokada/gh-gfm-preview/internal/assert"
)

func TestRenderHeadingsHTML(t *testing.T) {
	t.Run("extract headings and keep level", func(t *testing.T) {
		html := `<h1 id="main-title">Main Title</h1><p>x</p><h2 id="sub-title">Sub <code>Title</code></h2>`

		headingsHTML, hasHeadings := renderHeadingsHTML(html)

		assert.True(t, hasHeadings)
		assert.True(t, strings.Contains(headingsHTML, `href="#main-title"`))
		assert.True(t, strings.Contains(headingsHTML, `class="heading-item heading-level-1"`))
		assert.True(t, strings.Contains(headingsHTML, `>Main Title</a>`))
		assert.True(t, strings.Contains(headingsHTML, `href="#sub-title"`))
		assert.True(t, strings.Contains(headingsHTML, `class="heading-item heading-level-2"`))
		assert.True(t, strings.Contains(headingsHTML, `>Sub Title</a>`))
	})

	t.Run("escape heading text and id", func(t *testing.T) {
		html := `<h3 id='quoted"&id'>5 &lt; 8 &amp; "quoted"</h3>`

		headingsHTML, hasHeadings := renderHeadingsHTML(html)

		assert.True(t, hasHeadings)
		assert.True(t, strings.Contains(headingsHTML, `href="#quoted&#34;&amp;id"`))
		assert.True(t, strings.Contains(headingsHTML, `>5 &lt; 8 &amp; &#34;quoted&#34;</a>`))
	})

	t.Run("no headings", func(t *testing.T) {
		headingsHTML, hasHeadings := renderHeadingsHTML("<p>no headings</p>")

		assert.False(t, hasHeadings)
		assert.Equal(t, headingsHTML, "")
	})
}
