package cmd

import (
	"strings"
	"unicode"
)

// sanitizeField removes terminal control characters from implant- or
// provider-reported strings before they are printed to the operator's
// terminal. Without this, a value containing ANSI/CSI/OSC escape sequences
// (e.g. a self-reported callback host or user) could rewrite the operator's
// display or, on terminals that honor OSC 52, overwrite the clipboard.
//
// Any rune that is not printable — control characters below 0x20, DEL (0x7f),
// the C1 range 0x80-0x9f, and anything else failing unicode.IsPrint — is
// replaced with '?'.
func sanitizeField(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f) || !unicode.IsPrint(r) {
			return '?'
		}
		return r
	}, s)
}
