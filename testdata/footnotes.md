# Footnotes

GitHub flavored markdown should support footnotes.[^gfm]

This paragraph is intentionally left without footnotes so the first backlink
has nearby context to jump back into.

Repeated references should keep working.[^gfm]

Footnotes can also contain inline formatting like *emphasis* and `code`.[^formatting]

This is another paragraph without any references. It makes it easier to confirm
that backlink shortcuts return to the exact original reference instead of just
somewhere near the footnotes section.

The preview should also keep its scroll position stable while navigating
through multiple blocks of regular prose. This section exists only to create
distance between the references near the top of the file and the generated
footnotes near the bottom.

When checking this manually, it should be obvious whether the backlink from the
first footnote returns to the first reference or to one of the later repeated
references. A dense document makes that much easier to inspect.

The current fixture is intentionally verbose because short examples are fine for
parser tests but poor for navigation tests. Backlinks are primarily a browsing
behavior, so they need enough space on the page to be meaningful.

Another paragraph without any reference marker. It should remain visually
distinct from the reference paragraphs so the jump target is easier to spot.

Scrolling through a larger document also helps expose any accidental issues with
sticky headers, anchor offsets, or browser focus behavior. None of that should
change the generated HTML, but it affects how usable the final result feels.

To make the page longer, this fixture keeps alternating between plain prose and
occasional references. That pattern gives enough unique text around each anchor
so it is immediately clear which location the browser navigated back to.

This paragraph is still unreferenced. It exists to widen the gap between the
first and second clusters of footnote references.

Readers often skim long Markdown documents section by section. If footnote
backlinks land in the wrong place, that becomes more annoying as the distance
between the note and its source grows.

This fixture should therefore be long enough that the first reference is well
above the fold by the time the footnotes render at the bottom of the page.

Longer content also makes repeated-reference behavior easier to compare against
GitHub itself, because the `fnref`, `fnref1`, and later backlink destinations
become visually separated instead of collapsing into a single viewport.

The next few paragraphs are intentionally repetitive in structure but use
different wording so they are easy to distinguish during manual checks.

Plain body text can be boring, but for an anchor-navigation regression test it
is useful. What matters here is height, separation, and recognizable nearby
phrases rather than literary quality.

If a backlink from the first generated note lands here, that would be wrong.
This line gives a clear negative signal during a manual comparison.

If a backlink from the second generated note lands here, that would also be
wrong. The goal is to make incorrect jumps obvious immediately.

This fixture now keeps building vertical distance with regular paragraphs that
contain no special syntax at all.

That simplicity is deliberate: plain paragraphs reduce unrelated rendering
variables and keep the focus on footnote behavior.

One more unreferenced paragraph sits here so the first repeated footnote link is
not crowded by the later one.

Another plain paragraph extends the document further. At this point the
footnotes should already be meaningfully separated from the start of the file on
most screens.

This text is here so that manual testing on tall monitors is still useful. A
fixture that only spans one viewport does not adequately test backlink
navigation.

Some users may open the preview in a narrow split window. In that case the page
height grows faster, and the difference between correct and incorrect backlink
targets becomes even easier to observe.

For completeness, this paragraph has no reference either. It exists only to add
distance and unique wording.

The following repeated reference appears much later in the document than the
first two.[^gfm]

After that late reference, the file continues with several more ordinary
paragraphs so the reader can still scroll before reaching the note definitions.

This should help confirm that using the backlink from the footnote returns to
the exact late reference when selected from the corresponding back-arrow near
the footnote content.

Yet another plain paragraph pushes the final note definitions farther down.

Anchor navigation bugs often feel small in short test cases and obvious in long
ones. That is the entire point of this expanded fixture.

This sentence is only here to make the document longer.

This sentence is also only here to make the document longer.

This paragraph continues the same idea with slightly different wording so it is
not visually identical to the others.

Long prose gives each anchor a stronger local identity. That matters when you
are checking whether a browser jumped to the correct reference.

The fixture should now be long enough to require real scrolling in most preview
windows, but a few more paragraphs make the result more robust.

Here is more unreferenced content. Nothing special should happen here.

Here is another block of unreferenced content. It should simply render as a
normal paragraph.

Here is a final stretch of plain content before the last sentence with a
reference marker.

Much later in the document, the same note should still point back correctly.[^gfm]

The document ends with additional unreferenced text so the last reference is
not immediately adjacent to the generated footnotes.

This final paragraph should make it easier to verify whether a backlink lands on
the last reference or overshoots it.

[^gfm]: Footnotes are part of GitHub flavored markdown.
[^formatting]: Formatting inside footnotes should render normally.
