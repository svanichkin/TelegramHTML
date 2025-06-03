package telehtml

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/microcosm-cc/bluemonday"
)

func CleanTelegramHTML(html string) string {

	// SanitizeHTML

	html = sanitizeHTML(html)

	// Validate

	html = strings.ToValidUTF8(html, "")

	// Clean bed spaces, breake and etc

	html = strings.NewReplacer(
		"⠀", " ",
		"　", " ",
		" ", " ",
		" ", " ",
		"\u200c", " ",
		"\u00a0͏", " ",
		"\u00a0", " ",
		"\u034f", " ",
		"\t", "",
		"\r", "",
		"<br>", "\n",
		"<br />", "\n",
		"<br/>", "\n",
		"<p>", "\n",
		"</p>", "\n",
		"<strong>", "<b>",
		"</strong>", "</b>",
		"<em>", "<i>",
		"</em>", "</i>",
		"<strike>", "<s>",
		"</strike>", "</s>",
		"<del>", "<s>",
		"</del>", "</s>",
	).Replace(html)

	// Collapse 3+ spaces into a single newline

	html = regexp.MustCompile(` {3,}`).ReplaceAllString(html, "\n")

	// Trim inconsistent spacing around newlines

	html = strings.NewReplacer(
		"\n ", " ",
		" \n", "\n",
	).Replace(html)

	// Collapse multiple newlines into a single one

	reNewlines := regexp.MustCompile(`(\n[\s]*){2,}`)
	for {
		old := html
		html = reNewlines.ReplaceAllString(html, "\n")
		if html == old {
			break
		}
	}

	// Remove isolated single characters between newlines

	html = regexp.MustCompile(`\n.{1}\n`).ReplaceAllString(html, "\n")

	// Reapply newline collapsing to catch leftovers

	for {
		old := html
		html = reNewlines.ReplaceAllString(html, "\n")
		if html == old {
			break
		}
	}

	// Fix missing space before <a href> tags

	html = regexp.MustCompile(`([^\s\n])(<a href)`).ReplaceAllString(html, "$1 $2")

	// Fix missing space after </a> when followed by punctuation or word

	html = regexp.MustCompile(`</a>([^.,;!?:\s])`).ReplaceAllString(html, "</a> $1")

	// Add paragraph spacing before/after links and bold tags

	html = strings.NewReplacer(
		"\n<a href", "\n\n<a href",
		"</a>\n", "</a>\n\n",
		"\n<b", "\n\n<b",
		"</b>\n", "</b>\n\n",
	).Replace(html)

	// Clean excessive spaces after newlines

	html = regexp.MustCompile(`\n +`).ReplaceAllString(html, "\n")

	// Trim leading/trailing newlines and spaces

	html = strings.TrimLeft(html, "\n")
	html = strings.TrimRight(html, "\n")
	html = strings.TrimLeft(html, " ")
	html = strings.TrimRight(html, " ")

	return html
}

func sanitizeHTML(html string) string {

	// Sanitize tags

	p := bluemonday.NewPolicy()
	p.AllowElements("b", "strong", "i", "em", "u", "s", "strike", "del", "a", "code", "pre", "p", "br")
	p.AllowAttrs("href").OnElements("a")
	html = p.Sanitize(html)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return html
	}

	// Empty a href remove

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		trimmed := strings.TrimSpace(s.Text())
		if trimmed == "" && len(s.Children().Nodes) == 0 {
			s.Remove()
		} else {
			s.SetText(trimmed)
		}
	})

	// Create html

	html, err = doc.Html()
	if err != nil {
		return html
	}

	// Clean headers

	html = strings.TrimPrefix(html, "<html><head></head><body>")
	html = strings.TrimSuffix(html, "</body></html>")

	return html
}
