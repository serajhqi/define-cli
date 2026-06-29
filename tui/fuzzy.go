package tui

import "strings"

func levenshtein(a, b string) int {
	ra := []rune(a)
	rb := []rune(b)
	la := len(ra)
	lb := len(rb)

	d := make([][]int, la+1)
	for i := range d {
		d[i] = make([]int, lb+1)
		d[i][0] = i
	}
	for j := 0; j <= lb; j++ {
		d[0][j] = j
	}

	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			d[i][j] = min(
				d[i-1][j]+1,
				d[i][j-1]+1,
				d[i-1][j-1]+cost,
			)
		}
	}
	return d[la][lb]
}

func fuzzyMatch(filter, word string) bool {
	if filter == "" {
		return true
	}
	fl := strings.ToLower(filter)
	wl := strings.ToLower(word)

	if strings.HasPrefix(wl, fl) {
		return true
	}

	dist := levenshtein(fl, wl)
	return dist <= 2
}
