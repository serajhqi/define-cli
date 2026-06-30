package stem

import "strings"

func Stem(w string) string {
	w = strings.ToLower(w)
	if len(w) <= 2 {
		return w
	}

	w = step1a(w)
	w = step1b(w)
	w = step1c(w)
	w = step2(w)
	w = step3(w)
	w = step4(w)
	w = step5a(w)
	w = step5b(w)

	return w
}

func step1a(w string) string {
	if strings.HasSuffix(w, "sses") || strings.HasSuffix(w, "ies") {
		return w[:len(w)-2]
	}
	if strings.HasSuffix(w, "ss") {
		return w
	}
	if strings.HasSuffix(w, "s") {
		return w[:len(w)-1]
	}
	return w
}

func step1b(w string) string {
	if strings.HasSuffix(w, "eed") {
		if measure(w[:len(w)-3]) > 0 {
			return w[:len(w)-1]
		}
		return w
	}

	hasEd := strings.HasSuffix(w, "ed") && containsVowel(w[:len(w)-2])
	hasIng := strings.HasSuffix(w, "ing") && containsVowel(w[:len(w)-3])

	if !hasEd && !hasIng {
		return w
	}

	if hasEd {
		w = w[:len(w)-2]
	} else {
		w = w[:len(w)-3]
	}

	if strings.HasSuffix(w, "at") || strings.HasSuffix(w, "bl") || strings.HasSuffix(w, "iz") {
		return w + "e"
	}

	if len(w) >= 2 && w[len(w)-1] == w[len(w)-2] && strings.ContainsRune("bdgkmnprt", rune(w[len(w)-1])) {
		return w[:len(w)-1]
	}

	if measure(w) == 1 && endsWithCVCExceptWXY(w) {
		return w + "e"
	}

	return w
}

func step1c(w string) string {
	if strings.HasSuffix(w, "y") && len(w) >= 2 && containsVowel(w[:len(w)-1]) {
		return w[:len(w)-1] + "i"
	}
	return w
}

func step2(w string) string {
	suffixes := []struct{ from, to string }{
		{"ational", "ate"}, {"tional", "tion"}, {"enci", "ence"}, {"anci", "ance"},
		{"izer", "ize"}, {"abli", "able"}, {"alli", "al"}, {"entli", "ent"},
		{"eli", "e"}, {"ousli", "ous"}, {"ization", "ize"}, {"ation", "ate"},
		{"ator", "ate"}, {"alism", "al"}, {"iveness", "ive"}, {"fulness", "ful"},
		{"ousness", "ous"}, {"aliti", "al"}, {"iviti", "ive"}, {"biliti", "ble"},
	}
	for _, s := range suffixes {
		if strings.HasSuffix(w, s.from) && measure(w[:len(w)-len(s.from)]) > 0 {
			return w[:len(w)-len(s.from)] + s.to
		}
	}
	return w
}

func step3(w string) string {
	suffixes := []struct{ from, to string }{
		{"icate", "ic"}, {"ative", ""}, {"alize", "al"}, {"iciti", "ic"},
		{"ical", "ic"}, {"ful", ""}, {"ness", ""},
	}
	for _, s := range suffixes {
		if strings.HasSuffix(w, s.from) && measure(w[:len(w)-len(s.from)]) > 0 {
			return w[:len(w)-len(s.from)] + s.to
		}
	}
	return w
}

func step4(w string) string {
	suffixes := []string{"al", "ance", "ence", "er", "ic", "able", "ible", "ant", "ement", "ment", "ent", "ou", "ism", "ate", "iti", "ous", "ive", "ize"}
	for _, s := range suffixes {
		if strings.HasSuffix(w, s) && measure(w[:len(w)-len(s)]) > 1 {
			return w[:len(w)-len(s)]
		}
	}
	if strings.HasSuffix(w, "ion") && len(w) >= 4 && strings.ContainsRune("st", rune(w[len(w)-4])) && measure(w[:len(w)-3]) > 1 {
		return w[:len(w)-3]
	}
	return w
}

func step5a(w string) string {
	if strings.HasSuffix(w, "e") {
		stem := w[:len(w)-1]
		if measure(stem) > 1 {
			return stem
		}
		if measure(stem) == 1 && !endsWithCVCExceptWXY(stem) {
			return stem
		}
	}
	return w
}

func step5b(w string) string {
	if measure(w) > 1 && len(w) >= 2 && w[len(w)-1] == 'l' && w[len(w)-2] == 'l' {
		return w[:len(w)-1]
	}
	return w
}

func containsVowel(w string) bool {
	for _, c := range w {
		if isVowel(c) {
			return true
		}
	}
	return false
}

func isVowel(c rune) bool {
	return strings.ContainsRune("aeiou", c)
}

func measure(w string) int {
	n := 0
	inVowelSeq := false
	for _, c := range w {
		v := isVowel(c)
		if v && !inVowelSeq {
			inVowelSeq = true
		} else if !v && inVowelSeq {
			n++
			inVowelSeq = false
		}
	}
	return n
}

func endsWithCVCExceptWXY(w string) bool {
	if len(w) < 3 {
		return false
	}
	last := rune(w[len(w)-1])
	second := rune(w[len(w)-2])
	third := rune(w[len(w)-3])
	return !isVowel(third) && isVowel(second) && !isVowel(last) && last != 'w' && last != 'x' && last != 'y'
}

func Tokenize(s string) []string {
	var words []string
	var current strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			current.WriteRune(r)
		} else {
			if current.Len() > 0 {
				words = append(words, current.String())
				words = append(words, string(r))
				current.Reset()
			} else {
				words = append(words, string(r))
			}
		}
	}
	if current.Len() > 0 {
		words = append(words, current.String())
	}
	return words
}
