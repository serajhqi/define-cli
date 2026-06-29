package tui

import "testing"

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		a, b string
		dist int
	}{
		{"hello", "hello", 0},
		{"hello", "hallo", 1},
		{"hello", "hll", 2},
		{"abc", "", 3},
		{"", "abc", 3},
		{"kitten", "sitting", 3},
		{"flaw", "lawn", 2},
	}

	for _, tt := range tests {
		got := levenshtein(tt.a, tt.b)
		if got != tt.dist {
			t.Errorf("levenshtein(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.dist)
		}
	}
}

func TestFuzzyMatch(t *testing.T) {
	tests := []struct {
		filter string
		word   string
		match  bool
	}{
		{"hel", "hello", true},
		{"hllo", "hello", true},   // 1 typo
		{"hxllo", "hello", true},  // 1 extra char
		{"hexlo", "hello", true},  // 1 substitution
		{"hxlyo", "hello", true},  // Levenshtein distance 2
		{"zzz", "hello", false},   // distance 4
		{"", "hello", true},       // empty filter matches everything
		{"HEL", "hello", true},    // case-insensitive
	}

	for _, tt := range tests {
		got := fuzzyMatch(tt.filter, tt.word)
		if got != tt.match {
			t.Errorf("fuzzyMatch(%q, %q) = %v, want %v", tt.filter, tt.word, got, tt.match)
		}
	}
}

func TestFuzzyMatchPrefix(t *testing.T) {
	if !fuzzyMatch("sub", "submarine") {
		t.Error("prefix match should always pass")
	}
	if !fuzzyMatch("he", "hello") {
		t.Error("prefix match should pass")
	}
}
