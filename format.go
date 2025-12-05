package telegram

import (
	"strings"
)

// ParseMode constants for Telegram message formatting
const (
	ParseModeMarkdown   = "Markdown"
	ParseModeMarkdownV2 = "MarkdownV2"
	ParseModeHTML       = "HTML"
)

// EscapeMarkdownV2 escapes special characters for MarkdownV2 parse mode
// Characters that need escaping: _ * [ ] ( ) ~ ` > # + - = | { } . !
func EscapeMarkdownV2(text string) string {
	// Characters that must be escaped in MarkdownV2
	specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}

	result := text
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	return result
}

// EscapeHTML escapes special characters for HTML parse mode
func EscapeHTML(text string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
	)
	return replacer.Replace(text)
}

// Bold formats text as bold
func Bold(text string) string {
	return "*" + text + "*"
}

// BoldV2 formats text as bold for MarkdownV2 (escapes special chars in text)
func BoldV2(text string) string {
	return "*" + EscapeMarkdownV2(text) + "*"
}

// Italic formats text as italic
func Italic(text string) string {
	return "_" + text + "_"
}

// ItalicV2 formats text as italic for MarkdownV2 (escapes special chars in text)
func ItalicV2(text string) string {
	return "_" + EscapeMarkdownV2(text) + "_"
}

// Underline formats text as underline (MarkdownV2 only)
func Underline(text string) string {
	return "__" + text + "__"
}

// UnderlineV2 formats text as underline for MarkdownV2 (escapes special chars in text)
func UnderlineV2(text string) string {
	return "__" + EscapeMarkdownV2(text) + "__"
}

// Strikethrough formats text as strikethrough
func Strikethrough(text string) string {
	return "~" + text + "~"
}

// StrikethroughV2 formats text as strikethrough for MarkdownV2 (escapes special chars in text)
func StrikethroughV2(text string) string {
	return "~" + EscapeMarkdownV2(text) + "~"
}

// Spoiler formats text as spoiler (MarkdownV2 only)
func Spoiler(text string) string {
	return "||" + text + "||"
}

// SpoilerV2 formats text as spoiler for MarkdownV2 (escapes special chars in text)
func SpoilerV2(text string) string {
	return "||" + EscapeMarkdownV2(text) + "||"
}

// Code formats text as inline code
func Code(text string) string {
	return "`" + text + "`"
}

// CodeBlock formats text as code block
func CodeBlock(text string) string {
	return "```\n" + text + "\n```"
}

// CodeBlockWithLang formats text as code block with language
func CodeBlockWithLang(text, lang string) string {
	return "```" + lang + "\n" + text + "\n```"
}

// Link formats text as link
func Link(text, url string) string {
	return "[" + text + "](" + url + ")"
}

// LinkV2 formats text as link for MarkdownV2 (escapes special chars in text)
func LinkV2(text, url string) string {
	// Escape text but not URL (URL has different escaping rules)
	escapedText := EscapeMarkdownV2(text)
	// In URL, only ) and \ need escaping
	escapedURL := strings.ReplaceAll(url, "\\", "\\\\")
	escapedURL = strings.ReplaceAll(escapedURL, ")", "\\)")
	return "[" + escapedText + "](" + escapedURL + ")"
}

// Mention formats user mention
func Mention(text string, userID int64) string {
	return "[" + text + "](tg://user?id=" + formatInt64(userID) + ")"
}

// MentionV2 formats user mention for MarkdownV2
func MentionV2(text string, userID int64) string {
	return "[" + EscapeMarkdownV2(text) + "](tg://user?id=" + formatInt64(userID) + ")"
}

// BoldHTML formats text as bold in HTML
func BoldHTML(text string) string {
	return "<b>" + EscapeHTML(text) + "</b>"
}

// ItalicHTML formats text as italic in HTML
func ItalicHTML(text string) string {
	return "<i>" + EscapeHTML(text) + "</i>"
}

// UnderlineHTML formats text as underline in HTML
func UnderlineHTML(text string) string {
	return "<u>" + EscapeHTML(text) + "</u>"
}

// StrikethroughHTML formats text as strikethrough in HTML
func StrikethroughHTML(text string) string {
	return "<s>" + EscapeHTML(text) + "</s>"
}

// SpoilerHTML formats text as spoiler in HTML
func SpoilerHTML(text string) string {
	return "<tg-spoiler>" + EscapeHTML(text) + "</tg-spoiler>"
}

// CodeHTML formats text as inline code in HTML
func CodeHTML(text string) string {
	return "<code>" + EscapeHTML(text) + "</code>"
}

// CodeBlockHTML formats text as code block in HTML
func CodeBlockHTML(text string) string {
	return "<pre>" + EscapeHTML(text) + "</pre>"
}

// CodeBlockHTMLWithLang formats text as code block with language in HTML
func CodeBlockHTMLWithLang(text, lang string) string {
	return "<pre><code class=\"language-" + lang + "\">" + EscapeHTML(text) + "</code></pre>"
}

// LinkHTML formats text as link in HTML
func LinkHTML(text, url string) string {
	return "<a href=\"" + url + "\">" + EscapeHTML(text) + "</a>"
}

// MentionHTML formats user mention in HTML
func MentionHTML(text string, userID int64) string {
	return "<a href=\"tg://user?id=" + formatInt64(userID) + "\">" + EscapeHTML(text) + "</a>"
}

func formatInt64(n int64) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	if negative {
		digits = append([]byte{'-'}, digits...)
	}

	return string(digits)
}

// FormatMarkdownV2 processes text with markdown formatting
// Supports: *bold*, _italic_, `code`, ```pre```, [link](url), ~strikethrough~, __underline__, ||spoiler||
// Escapes special characters outside of formatting blocks
func FormatMarkdownV2(text string) string {
	if text == "" {
		return ""
	}

	var result strings.Builder
	runes := []rune(text)
	i := 0

	for i < len(runes) {
		// Check for code block ```
		if i+2 < len(runes) && runes[i] == '`' && runes[i+1] == '`' && runes[i+2] == '`' {
			end := findClosingCodeBlock(runes, i+3)
			if end != -1 {
				result.WriteString(string(runes[i : end+3]))
				i = end + 3
				continue
			}
		}

		// Check for inline code `
		if runes[i] == '`' {
			end := findClosingChar(runes, i+1, '`')
			if end != -1 {
				result.WriteString(string(runes[i : end+1]))
				i = end + 1
				continue
			}
		}

		// Check for spoiler ||
		if i+1 < len(runes) && runes[i] == '|' && runes[i+1] == '|' {
			end := findClosingDouble(runes, i+2, '|')
			if end != -1 {
				// Content inside spoiler needs escaping
				content := escapeInsideFormat(string(runes[i+2 : end]))
				result.WriteString("||")
				result.WriteString(content)
				result.WriteString("||")
				i = end + 2
				continue
			}
		}

		// Check for underline __
		if i+1 < len(runes) && runes[i] == '_' && runes[i+1] == '_' {
			end := findClosingDouble(runes, i+2, '_')
			if end != -1 {
				content := escapeInsideFormat(string(runes[i+2 : end]))
				result.WriteString("__")
				result.WriteString(content)
				result.WriteString("__")
				i = end + 2
				continue
			}
		}

		// Check for bold *
		if runes[i] == '*' {
			end := findClosingChar(runes, i+1, '*')
			if end != -1 {
				content := escapeInsideFormat(string(runes[i+1 : end]))
				result.WriteRune('*')
				result.WriteString(content)
				result.WriteRune('*')
				i = end + 1
				continue
			}
		}

		// Check for italic _
		if runes[i] == '_' && (i+1 >= len(runes) || runes[i+1] != '_') {
			end := findClosingChar(runes, i+1, '_')
			if end != -1 && (end+1 >= len(runes) || runes[end+1] != '_') {
				content := escapeInsideFormat(string(runes[i+1 : end]))
				result.WriteRune('_')
				result.WriteString(content)
				result.WriteRune('_')
				i = end + 1
				continue
			}
		}

		// Check for strikethrough ~
		if runes[i] == '~' {
			end := findClosingChar(runes, i+1, '~')
			if end != -1 {
				content := escapeInsideFormat(string(runes[i+1 : end]))
				result.WriteRune('~')
				result.WriteString(content)
				result.WriteRune('~')
				i = end + 1
				continue
			}
		}

		// Check for link [text](url)
		if runes[i] == '[' {
			linkEnd := parseLinkMarkdown(runes, i)
			if linkEnd != -1 {
				result.WriteString(string(runes[i : linkEnd+1]))
				i = linkEnd + 1
				continue
			}
		}

		// Escape regular character if it's special
		if isMarkdownV2Special(runes[i]) {
			result.WriteRune('\\')
		}
		result.WriteRune(runes[i])
		i++
	}

	return result.String()
}

// findClosingCodeBlock finds closing ``` for code block
func findClosingCodeBlock(runes []rune, start int) int {
	for i := start; i+2 < len(runes); i++ {
		if runes[i] == '`' && runes[i+1] == '`' && runes[i+2] == '`' {
			return i
		}
	}
	return -1
}

// findClosingChar finds closing character
func findClosingChar(runes []rune, start int, char rune) int {
	for i := start; i < len(runes); i++ {
		if runes[i] == char {
			return i
		}
		// Skip escaped characters
		if runes[i] == '\\' && i+1 < len(runes) {
			i++
		}
	}
	return -1
}

// findClosingDouble finds closing double character (||, __)
func findClosingDouble(runes []rune, start int, char rune) int {
	for i := start; i+1 < len(runes); i++ {
		if runes[i] == char && runes[i+1] == char {
			return i
		}
		// Skip escaped characters
		if runes[i] == '\\' && i+1 < len(runes) {
			i++
		}
	}
	return -1
}

// parseLinkMarkdown parses [text](url) and returns end index
func parseLinkMarkdown(runes []rune, start int) int {
	if runes[start] != '[' {
		return -1
	}

	// Find ]
	bracketEnd := -1
	for i := start + 1; i < len(runes); i++ {
		if runes[i] == ']' {
			bracketEnd = i
			break
		}
		if runes[i] == '\\' && i+1 < len(runes) {
			i++
		}
	}

	if bracketEnd == -1 || bracketEnd+1 >= len(runes) || runes[bracketEnd+1] != '(' {
		return -1
	}

	// Find )
	parenEnd := -1
	depth := 1
	for i := bracketEnd + 2; i < len(runes); i++ {
		if runes[i] == '(' {
			depth++
		} else if runes[i] == ')' {
			depth--
			if depth == 0 {
				parenEnd = i
				break
			}
		}
		if runes[i] == '\\' && i+1 < len(runes) {
			i++
		}
	}

	return parenEnd
}

// escapeInsideFormat escapes special chars inside formatting blocks
// Does not escape the formatting character itself
func escapeInsideFormat(text string) string {
	// Inside formatted text, we need to escape: ) ( ` \ and >
	// but NOT the formatting chars themselves
	specialInside := []string{"\\", "`", ")", "(", ">"}
	result := text
	for _, char := range specialInside {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	return result
}

// isMarkdownV2Special checks if rune is a special MarkdownV2 character
func isMarkdownV2Special(r rune) bool {
	switch r {
	case '_', '*', '[', ']', '(', ')', '~', '`', '>', '#', '+', '-', '=', '|', '{', '}', '.', '!', '\\':
		return true
	}
	return false
}

// StripMarkdown removes all markdown formatting from text
func StripMarkdown(text string) string {
	// Remove code blocks
	result := text

	// Simple removal of formatting characters
	for _, char := range []string{"```", "||", "__", "*", "_", "~", "`"} {
		result = strings.ReplaceAll(result, char, "")
	}

	// Remove escape characters
	result = strings.ReplaceAll(result, "\\", "")

	return result
}

// TruncateText truncates text to maxLen, adding "..." if truncated
// Respects UTF-8 runes
func TruncateText(text string, maxLen int) string {
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}
