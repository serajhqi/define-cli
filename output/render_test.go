package output_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/seraj/define/api"
	"github.com/seraj/define/output"
)

var ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string { return ansiRE.ReplaceAllString(s, "") }

func TestRenderDefinition(t *testing.T) {
	def := &api.Definition{
		Word:      "aim",
		Phonetics: []api.Phonetic{{Text: "/eɪm/"}},
		Meanings: []api.Meaning{
			{
				PartOfSpeech: "verb",
				Definitions: []api.Def{
					{Definition: "To point or direct a weapon at a target.", Example: "He aimed the gun at the target."},
					{Definition: "To strive or intend to do something.", Example: "I aim to finish this today."},
				},
			},
			{
				PartOfSpeech: "noun",
				Definitions: []api.Def{
					{Definition: "A purpose or goal; an objective.", Example: "His aim was to become a doctor."},
				},
			},
		},
	}

	result := output.Render(def)
	plain := stripANSI(result)
	lines := strings.Split(plain, "\n")

	// Header: word and phonetic
	if !strings.Contains(lines[0], "aim") {
		t.Errorf("first line should contain word, got: %q", lines[0])
	}
	if !strings.Contains(strings.Join(lines[:3], "\n"), "/eɪm/") {
		t.Error("should contain phonetic")
	}

	// Parts of speech
	full := result
	if !strings.Contains(full, "verb") {
		t.Error("should contain verb section")
	}
	if !strings.Contains(full, "noun") {
		t.Error("should contain noun section")
	}

	// Definitions
	if !strings.Contains(full, "point or direct") {
		t.Error("should contain first definition")
	}
	if !strings.Contains(full, "purpose or goal") {
		t.Error("should contain noun definition")
	}

	// Examples
	if !strings.Contains(full, "He aimed the gun") {
		t.Error("should contain first example")
	}
	if !strings.Contains(full, "I aim to finish") {
		t.Error("should contain second example")
	}

	// Width check: no line should exceed 70 chars
	for i, line := range lines {
		if len(line) > 70 {
			t.Errorf("line %d exceeds 70 chars: %d chars: %q", i, len(line), line)
		}
	}
}

func TestRenderError(t *testing.T) {
	result := output.RenderError("serendipity", "not found in dictionary")
	plain := stripANSI(result)
	if !strings.Contains(plain, "serendipity") {
		t.Error("should contain the word")
	}
	if !strings.Contains(plain, "not found") {
		t.Error("should contain error message")
	}
	if !strings.Contains(plain, "✖") {
		t.Error("should contain error marker")
	}
}
