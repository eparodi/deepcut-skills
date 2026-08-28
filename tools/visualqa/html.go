package main

import (
	"regexp"
)

// htmlTruncatedMarker is appended when the prepared HTML is capped so the
// model knows the structure it receives is partial (US8 AC2).
const htmlTruncatedMarker = "\n…[truncated]"

var (
	reScript  = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	reStyle   = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	reComment = regexp.MustCompile(`(?s)<!--.*?-->`)
)

// sanitizeHTML strips script/style blocks and HTML comments. Best-effort
// regex hygiene: the goal is token-cost reduction and prompt-noise removal,
// NOT a security boundary (the HTML comes from the page under test and never
// executes anything here).
func sanitizeHTML(raw string) string {
	s := reScript.ReplaceAllString(raw, "")
	s = reStyle.ReplaceAllString(s, "")
	return reComment.ReplaceAllString(s, "")
}

// prepareHTML sanitizes and caps the page HTML for the vision request.
// maxChars <= 0 means no cap. A capped payload gets the explicit truncation
// marker so the model never mistakes a partial DOM for the whole page.
func prepareHTML(raw string, maxChars int) string {
	clean := sanitizeHTML(raw)
	if maxChars <= 0 || len(clean) <= maxChars {
		return clean
	}
	return clean[:maxChars] + htmlTruncatedMarker
}
