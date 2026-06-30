package stem_test

import (
	"testing"

	"github.com/seraj/define/stem"
)

func TestPorterStem(t *testing.T) {
	tests := []struct{ word, stem string }{
		{"aiming", "aim"},
		{"aimed", "aim"},
		{"aims", "aim"},
		{"running", "run"},
		{"tests", "test"},
		{"testing", "test"},
		{"tested", "test"},
		{"happy", "happi"},
		{"happiness", "happi"},
		{"relational", "relat"},
		{"conditional", "condit"},
		{"generalization", "gener"},
		{"national", "nation"},
	}

	for _, tt := range tests {
		got := stem.Stem(tt.word)
		if got != tt.stem {
			t.Errorf("Stem(%q) = %q, want %q", tt.word, got, tt.stem)
		}
	}
}

func TestStemHighlightsInflections(t *testing.T) {
	s := stem.Stem("running")
	if s != "run" {
		t.Fatalf("stem of running should be run, got %q", s)
	}

	s2 := stem.Stem("run")
	if s2 != "run" {
		t.Fatalf("stem of run should be run, got %q", s2)
	}
}

func TestStemNoFalsePositive(t *testing.T) {
	stemAim := stem.Stem("aim")
	stemClaim := stem.Stem("claim")

	if stemAim == stemClaim {
		t.Errorf("aim (%q) and claim (%q) should not have the same stem", stemAim, stemClaim)
	}
}
