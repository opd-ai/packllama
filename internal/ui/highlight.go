package ui

import (
	"image/color"
	"strings"
)

// textSegment is a contiguous run of text in a message, optionally a code block.
type textSegment struct {
	text   string
	isCode bool
	lang   string // language hint after the opening fence, may be empty
}

// parseSegments splits content on ``` code-fence markers.
// Segments alternate between prose text and code blocks.
// Unclosed fences are treated as the rest of the string being a code block.
func parseSegments(content string) []textSegment {
	var segs []textSegment
	inCode := false
	lang := ""
	var buf strings.Builder
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			segs = append(segs, flushSegment(buf.String(), inCode, lang))
			buf.Reset()
			if !inCode {
				lang = strings.TrimPrefix(trimmed, "```")
			} else {
				lang = ""
			}
			inCode = !inCode
			continue
		}
		buf.WriteString(line)
		buf.WriteByte('\n')
	}
	if buf.Len() > 0 {
		segs = append(segs, flushSegment(buf.String(), inCode, lang))
	}
	return segs
}

// flushSegment creates a segment from buffered text, trimming trailing newlines.
func flushSegment(text string, isCode bool, lang string) textSegment {
	return textSegment{
		text:   strings.TrimRight(text, "\n"),
		isCode: isCode,
		lang:   lang,
	}
}

// segmentLines converts a segment into display lines annotated with code metadata.
func segmentLines(seg textSegment) []displayLine {
	raw := strings.Split(seg.text, "\n")
	lines := make([]displayLine, 0, len(raw))
	for _, l := range raw {
		lines = append(lines, displayLine{text: l, isCode: seg.isCode})
	}
	return lines
}

// displayLine is one rendered row in the chat view.
type displayLine struct {
	text   string
	isCode bool
	msgIdx int // index into the message list; -1 for separators
}

// langKeywords holds reserved words for common languages.
var langKeywords = map[string][]string{
	"go":         {"func", "var", "const", "type", "struct", "interface", "import", "package", "return", "if", "else", "for", "range", "switch", "case", "default", "go", "defer", "chan", "map", "make", "new", "nil", "true", "false"},
	"python":     {"def", "class", "import", "from", "return", "if", "elif", "else", "for", "while", "in", "not", "and", "or", "True", "False", "None", "with", "as", "pass", "break", "continue", "lambda", "yield"},
	"javascript": {"function", "var", "let", "const", "return", "if", "else", "for", "while", "class", "import", "export", "default", "new", "this", "null", "undefined", "true", "false", "async", "await"},
	"typescript": {"function", "var", "let", "const", "return", "if", "else", "for", "while", "class", "interface", "type", "import", "export", "default", "new", "this", "null", "undefined", "true", "false", "async", "await"},
}

// syntaxColors maps token kinds to display colors.
var syntaxColors = struct {
	keyword color.Color
	string_ color.Color
	comment color.Color
	number  color.Color
}{
	keyword: color.RGBA{R: 0xcb, G: 0xa6, B: 0xf7, A: 0xff}, // purple
	string_: color.RGBA{R: 0xa6, G: 0xe3, B: 0xa1, A: 0xff}, // green
	comment: color.RGBA{R: 0x6c, G: 0x70, B: 0x86, A: 0xff}, // grey
	number:  color.RGBA{R: 0xfa, G: 0xb3, B: 0x87, A: 0xff}, // orange
}

// HighlightLine tokenises line for the given language and returns colored spans.
// Falls back to plain text when the language is unrecognised.
func HighlightLine(line, lang string) []coloredSpan {
	if isCommentLine(line, lang) {
		return []coloredSpan{{text: line, clr: syntaxColors.comment}}
	}
	if isStringLine(line) {
		return []coloredSpan{{text: line, clr: syntaxColors.string_}}
	}
	keywords := langKeywords[strings.ToLower(lang)]
	if len(keywords) == 0 {
		return []coloredSpan{{text: line, clr: color.White}}
	}
	return tokeniseLine(line, keywords)
}

// isCommentLine reports whether line is a single-line comment.
func isCommentLine(line, lang string) bool {
	trimmed := strings.TrimSpace(line)
	switch strings.ToLower(lang) {
	case "go", "javascript", "typescript":
		return strings.HasPrefix(trimmed, "//")
	case "python":
		return strings.HasPrefix(trimmed, "#")
	}
	return strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#")
}

// isStringLine reports whether the line appears to be a pure string literal.
func isStringLine(line string) bool {
	t := strings.TrimSpace(line)
	return (strings.HasPrefix(t, `"`) && strings.HasSuffix(t, `"`)) ||
		(strings.HasPrefix(t, `'`) && strings.HasSuffix(t, `'`))
}

// tokeniseLine splits line into keyword and non-keyword spans.
func tokeniseLine(line string, keywords []string) []coloredSpan {
	var spans []coloredSpan
	remaining := line
	for remaining != "" {
		word, before, after, found := extractWord(remaining, keywords)
		if !found {
			spans = append(spans, coloredSpan{text: remaining, clr: color.White})
			break
		}
		if before != "" {
			spans = append(spans, coloredSpan{text: before, clr: color.White})
		}
		spans = append(spans, coloredSpan{text: word, clr: syntaxColors.keyword})
		remaining = after
	}
	return spans
}

// extractWord finds the first keyword in s, returning the text before, the
// keyword itself, the text after, and whether a keyword was found.
func extractWord(s string, keywords []string) (word, before, after string, found bool) {
	earliest := len(s) + 1
	for _, kw := range keywords {
		idx := strings.Index(s, kw)
		if idx < 0 || idx >= earliest {
			continue
		}
		// Ensure it's a whole word (not part of a longer identifier).
		end := idx + len(kw)
		if idx > 0 && isIdentChar(rune(s[idx-1])) {
			continue
		}
		if end < len(s) && isIdentChar(rune(s[end])) {
			continue
		}
		earliest = idx
		word = kw
		found = true
	}
	if !found {
		return "", s, "", false
	}
	before = s[:earliest]
	after = s[earliest+len(word):]
	return word, before, after, true
}

// isIdentChar reports whether r is a letter, digit, or underscore.
func isIdentChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}
