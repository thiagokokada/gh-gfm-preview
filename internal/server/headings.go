package server

import (
	stdhtml "html"
	"regexp"
	"strconv"
	"strings"
)

type headingItem struct {
	Level int
	ID    string
	Text  string
}

var (
	headingTagRegexp = regexp.MustCompile(`(?is)<h([1-6])([^>]*)>(.*?)</h([1-6])>`)
	headingIDRegexp  = regexp.MustCompile(`(?is)\bid\s*=\s*(?:"([^"]*)"|'([^']*)')`)
	htmlTagRegexp    = regexp.MustCompile(`(?is)<[^>]+>`)
)

func renderHeadingsHTML(markdownHTML string) (string, bool) {
	headings := extractHeadingItems(markdownHTML)
	if len(headings) == 0 {
		return "", false
	}

	var builder strings.Builder

	for _, heading := range headings {
		builder.WriteString(`<a href="#`)
		builder.WriteString(stdhtml.EscapeString(heading.ID))
		builder.WriteString(`" class="heading-item heading-level-`)
		builder.WriteString(strconv.Itoa(heading.Level))
		builder.WriteString(`">`)
		builder.WriteString(stdhtml.EscapeString(heading.Text))
		builder.WriteString(`</a>`)
	}

	return builder.String(), true
}

func extractHeadingItems(markdownHTML string) []headingItem {
	matches := headingTagRegexp.FindAllStringSubmatch(markdownHTML, -1)
	if len(matches) == 0 {
		return nil
	}

	headings := make([]headingItem, 0, len(matches))

	for _, match := range matches {
		if match[1] != match[4] {
			continue
		}

		level, err := strconv.Atoi(match[1])
		if err != nil {
			continue
		}

		id := extractHeadingID(match[2])
		if id == "" {
			continue
		}

		text := extractHeadingText(match[3])
		if text == "" {
			continue
		}

		headings = append(headings, headingItem{
			Level: level,
			ID:    id,
			Text:  text,
		})
	}

	return headings
}

func extractHeadingID(attributes string) string {
	match := headingIDRegexp.FindStringSubmatch(attributes)
	if len(match) < 3 {
		return ""
	}

	if match[1] != "" {
		return stdhtml.UnescapeString(match[1])
	}

	return stdhtml.UnescapeString(match[2])
}

func extractHeadingText(innerHTML string) string {
	text := htmlTagRegexp.ReplaceAllString(innerHTML, "")
	text = stdhtml.UnescapeString(text)
	text = strings.Join(strings.Fields(text), " ")

	return strings.TrimSpace(text)
}
