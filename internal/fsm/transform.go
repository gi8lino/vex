package fsm

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// transformCase applies ^,^^ (upper) and ,,, ,, (lower) on runes.
func transformCase(op, s string) string {
	switch op {
	case "^^":
		return strings.ToUpper(s)
	case ",,":
		return strings.ToLower(s)
	case "^":
		if s == "" {
			return s
		}
		r, size := utf8.DecodeRuneInString(s)
		r = unicode.ToUpper(r)
		return string(r) + s[size:]
	case ",":
		if s == "" {
			return s
		}
		r, size := utf8.DecodeRuneInString(s)
		r = unicode.ToLower(r)
		return string(r) + s[size:]
	default:
		return s
	}
}

// substr parses "off[:len]" and slices by rune index.
// Negative off counts from the end; len is optional.
func substr(spec, s string) string {
	off, length, hasLen := parseOffsetLen(spec)
	rs := []rune(s) // Simpler/clearer; switch to a zero-copy utf8 version if profiling demands it.
	n := len(rs)

	// Resolve start index i (clamped to [0, n]).
	i := off
	if off < 0 {
		i = n + off // count from end if negative
	}
	i = min(i, n) // clamp
	i = max(0, i) // floor at start

	// Resolve end index j (clamped to [i, n]).
	j := n
	if hasLen {
		j = i + length
		j = min(j, n) // cap at end
		j = max(j, i) // don’t go before start
	}

	return string(rs[i:j])
}

// parseOffsetLen parses "off[:len]" and returns (off, len, hasLen).
// It is tolerant: non-digits stop parsing; leading '-' sets the sign.
func parseOffsetLen(spec string) (off int, length int, hasLen bool) {
	left, right, ok := strings.Cut(spec, ":")
	off = atoiSafe(left)
	if ok {
		length = atoiSafe(right)
		hasLen = true
	}
	return
}

// atoiSafe parses a possibly-signed decimal prefix of s into int.
// Non-digits stop parsing; empty/invalid returns 0.
func atoiSafe(s string) int {
	sign := 1
	if strings.HasPrefix(s, "-") {
		sign = -1
		s = s[1:]
	}
	n := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return sign * n
}

// trimPrefixAll removes repeated matching prefixes until none remain.
func trimPrefixAll(s, prefix string) string {
	if prefix == "" {
		return s
	}
	for {
		var ok bool
		s, ok = strings.CutPrefix(s, prefix)
		if !ok {
			return s
		}
	}
}

// trimSuffixAll removes repeated matching suffixes until none remain.
func trimSuffixAll(s, suffix string) string {
	if suffix == "" {
		return s
	}
	for {
		var ok bool
		s, ok = strings.CutSuffix(s, suffix)
		if !ok {
			return s
		}
	}
}
