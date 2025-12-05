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
