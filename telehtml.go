package telehtml

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/microcosm-cc/bluemonday"
)

// Cleaner

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

// Splitter

const maxLen = 4000

func SplitTelegramHTML(html string) []string {

	var blocks []string
	for len(html) > 0 {

		// If text length small - no need split

		if len(html) < maxLen {
			blocks = append(blocks, html)
			break
		}

		// Cut text

		cut := cutTextBeforePos(html, maxLen)

		//  Find best cut point

		cut, pos, open := findCutPoint(cut)

		// Add completed message block

		blocks = append(blocks, cut)

		// Open closed tag, if needed and cut text

		html = open + cutTextAfterPos(html, pos)
	}

	return blocks
}

func findCutPoint(cut string) (string, int, string) {

	// Find last cut point with tag

	positions := []int{
		findLastTagRightPosition(cut, "</a>", "\n"),
		findLastTagRightPosition(cut, "</b>", ""),
		findLastTagRightPosition(cut, "</code>", ""),
		findLastTagRightPosition(cut, "</i>", ""),
		findLastTagRightPosition(cut, "</pre>", ""),
		findLastTagRightPosition(cut, "</s>", ""),
		findLastTagRightPosition(cut, "</u>", ""),
		findLastTagRightPosition(cut, "\n", ""),
	}
	maxPos := -1
	for _, pos := range positions {
		if pos > maxPos {
			maxPos = pos
		}
	}

	// if notfound, cut with bad point

	if maxPos == -1 {
		maxPos = findLastTagLeftPosition(cut, "<a href=")
	}
	if maxPos == -1 {
		maxPos = findLastTagRightPosition(cut, ". ", "")
	}
	if maxPos == -1 {
		maxPos = findLastTagRightPosition(cut, ", ", "")
	}

	// if not found, cut with length

	if maxPos == -1 {
		maxPos = len(cut)
	}

	// Cut text

	cut = cutTextBeforePos(cut, maxPos)

	// Check for Close tag and close if needed

	open, close := findEnclosingTags(cut, maxPos)
	if len(close) > 0 {
		cut = cut + close
	}

	return cut, maxPos, open
}

func cutTextBeforePos(text string, pos int) string {

	if pos > len(text) {
		pos = len(text)
	}
	if pos < 0 {
		pos = 0
	}

	return text[:pos]
}

func cutTextAfterPos(text string, pos int) string {

	if pos < 0 {
		pos = 0
	}
	if pos > len(text) {
		pos = len(text)
	}

	return text[pos:]
}

func findLastTagRightPosition(text, prefix, postfix string) int {

	lastPos := -1
	offset := 0
	for {

		// Main work

		idx := strings.Index(text[offset:], prefix)
		if idx == -1 {
			break
		}
		absoluteIdx := offset + idx
		pos := absoluteIdx + len(prefix)

		// Postfix work

		hasPostfix := false
		if postfix == "" {
			hasPostfix = true
		} else if pos <= len(text)-len(postfix) && strings.HasPrefix(text[pos:], postfix) {
			hasPostfix = true
		}
		if hasPostfix {
			lastPos = pos
		}

		// New offset

		offset = absoluteIdx + 1
	}

	return lastPos
}

func findLastTagLeftPosition(text, prefix string) int {

	lastPos := -1
	offset := 0
	for {
		idx := strings.Index(text[offset:], prefix)
		if idx == -1 {
			break
		}
		absoluteIdx := offset + idx
		lastPos = absoluteIdx
		offset = absoluteIdx + 1
	}

	return lastPos
}

func findEnclosingTags(text string, pos int) (string, string) {

	if pos > len(text) {
		pos = len(text)
	}
	i := pos - 1
	for i >= 0 {
		if text[i] == '<' {

			// Search start tag

			end := i + 1
			for end < len(text) && text[end] != '>' {
				end++
			}
			if end >= len(text) {
				break
			}
			tagContent := text[i+1 : end]
			tagParts := strings.Fields(tagContent)
			if len(tagParts) == 0 {
				break
			}
			tagName := tagParts[0]

			// If closed tag - skip

			if strings.HasPrefix(tagName, "/") {
				return "", ""
			}
			tagNameClean := strings.Split(tagName, " ")[0]
			openTag := "<" + tagName + ">"
			closeTag := "</" + tagNameClean + ">"
			return openTag, closeTag
		}
		i--
	}

	return "", ""
}

// Invisible Int tag

func EncodeIntInvisible(n int) string {
	
	if n == 0 {
		return string(invisibleRunes[0])
	}
	
	var result []rune
	for n > 0 {
		digit := n % 10
		result = append([]rune{invisibleRunes[digit]}, result...)
		n /= 10
	}
	
	return string(result)
}

func DecodeIntInvisible(s string) int {
	
	var result int
	for _, r := range s {
		if d, ok := runeToDigit[r]; ok {
			result = result*10 + d
		}
	}
	
	return result
}

var invisibleRunes = []rune{
	'\u200B', // 0 Zero Width Space
	'\u200C', // 1 Zero Width Non-Joiner
	'\u200D', // 2 Zero Width Joiner
	'\u2060', // 3 Word Joiner
	'\uFEFF', // 4 Zero Width No-Break Space
	'\u2061', // 5 Function Application
	'\u2062', // 6 Invisible Times
	'\u2063', // 7 Invisible Separator
	'\u2064', // 8 Invisible Plus
	'\u034F', // 9 Combining Grapheme Joiner
}

var runeToDigit = func() map[rune]int {
	
	m := make(map[rune]int)
	for i, r := range invisibleRunes {
		m[r] = int(i)
	}
	return m
	
}()

var invisibleSet = func() map[rune]bool {
	
	m := make(map[rune]bool)
	for _, r := range invisibleRunes {
		m[r] = true
	}
	return m
	
}()

func findInvisibleUidSequences(s string) []string {
	
	var sequences []string
	var current []rune
	
	for _, r := range s {
		if invisibleSet[r] {
			current = append(current, r)
		} else if len(current) > 0 {
			sequences = append(sequences, string(current))
			current = nil
		}
	}
	if len(current) > 0 {
		sequences = append(sequences, string(current))
	}
	
	return sequences
}
