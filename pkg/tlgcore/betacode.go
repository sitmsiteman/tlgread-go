package tlgcore

import (
	"bytes"
	"regexp"
	"strings"
	"unicode"
)

var greekBase = map[rune]rune{
	'a': 'α', 'b': 'β', 'g': 'γ', 'd': 'δ', 'e': 'ε', 'z': 'ζ', 'h': 'η', 'q': 'θ',
	'i': 'ι', 'k': 'κ', 'l': 'λ', 'm': 'μ', 'n': 'ν', 'c': 'ξ', 'o': 'ο', 'p': 'π',
	'r': 'ρ', 's': 'σ', 'j': 'ς', 't': 'τ', 'u': 'υ', 'f': 'φ', 'x': 'χ', 'y': 'ψ', 'w': 'ω',
}

var diacritics = map[rune]string{
	')': "\u0313", '(': "\u0314", '/': "\u0301", '\\': "\u0300",
	'=': "\u0342", '+': "\u0308", '|': "\u0345",
}

func ToGreek(s string) string {
	var out bytes.Buffer
	upper := false
	
	// Cleanup artifacts
	s = regexp.MustCompile(`@\{.*?\}`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`[\[\]%$]\d*`).ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, "6", "")
	s = strings.ReplaceAll(s, "1", "")

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '*' {
			upper = true
			continue
		}
		if c, ok := greekBase[unicode.ToLower(r)]; ok {
			if upper {
				out.WriteRune(unicode.ToUpper(c))
				upper = false
			} else {
				out.WriteRune(c)
			}
		} else if d, ok := diacritics[r]; ok {
			out.WriteString(d)
		} else {
			out.WriteRune(r)
		}
	}
	res := out.String()
	res = strings.ReplaceAll(res, "><", ">>")
	res = regexp.MustCompile(`σ(\s|[[:punct:]]|$)`).ReplaceAllString(res, "ς$1")
	return res
}

func ToBetaCode(s string) string {
	var out strings.Builder
	m := map[rune]string{
		'α': "a", 'β': "b", 'γ': "g", 'δ': "d", 'ε': "e", 'ζ': "z", 'η': "h", 'θ': "q",
		'ι': "i", 'κ': "k", 'λ': "l", 'μ': "m", 'ν': "n", 'ξ': "c", 'ο': "o", 'π': "p",
		'ρ': "r", 'σ': "s", 'ς': "s", 'τ': "t", 'υ': "u", 'φ': "f", 'χ': "x", 'ψ': "y", 'ω': "w",
		'ά': "a/", 'έ': "e/", 'ή': "h/", 'ί': "i/", 'ό': "o/", 'ύ': "u/", 'ώ': "w/",
	}
	for _, r := range s {
		lower := unicode.ToLower(r)
		if val, ok := m[lower]; ok {
			out.WriteString(val)
		} else {
			out.WriteRune(r)
		}
	}
	return out.String()
}

