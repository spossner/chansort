// --- path: internal/scm/strings_go118.go ---
//go:build go1.18
// +build go1.18

package scm

import "strings"

func equalFold(a, b string) bool { return strings.EqualFold(a, b) }
func hasSuffixFold(s, suf string) bool {
	if len(suf) == 0 {
		return false
	}
	if len(s) < len(suf) {
		return false
	}
	return strings.EqualFold(s[len(s)-len(suf):], suf)
}
