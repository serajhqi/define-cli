package output

import (
	"fmt"
	"strings"

	"github.com/seraj/define/api"
)

const width = 70

const (
	reset        = "\033[0m"
	bold         = "\033[1m"
	dim          = "\033[2m"
	white        = "\033[37m"
	brightYellow = "\033[93m"
	cyan         = "\033[36m"
	red          = "\033[91m"
)

func wordHeader(s string) string     { return bold + brightYellow + s + reset }
func posLabel(s string) string       { return bold + cyan + s + reset }
func definitionText(s string) string { return white + s + reset }
func exampleText(s string) string    { return dim + s + reset }
func treeLine(s string) string       { return dim + s + reset }
func errorText(s string) string      { return red + s + reset }

func Render(def *api.Definition) string {
	var b strings.Builder

	b.WriteString("  " + wordHeader(def.Word))
	if len(def.Phonetics) > 0 {
		b.WriteString("  " + wordHeader(def.Phonetics[0].Text))
	}
	b.WriteString("\n\n")

	for mi, meaning := range def.Meanings {
		isLastPOS := mi == len(def.Meanings)-1

		if isLastPOS {
			b.WriteString("  " + treeLine("╰─▸") + " ")
		} else {
			b.WriteString("  " + treeLine("├─▸") + " ")
		}
		b.WriteString(posLabel(meaning.PartOfSpeech))
		b.WriteString("\n")

		for di, def := range meaning.Definitions {
			var prefix string
			if isLastPOS {
				prefix = "      "
			} else {
				prefix = "  │   "
			}

			num := fmt.Sprintf("%d. ", di+1)
			line := treeLine(prefix+"├─ ") + definitionText(num+def.Definition)
			if plainLen(line) > width {
				line = truncateColored(line, width)
			}
			b.WriteString(line)
			b.WriteString("\n")

			if def.Example != "" {
				var exPrefix string
				if isLastPOS {
					exPrefix = "      │    "
				} else {
					exPrefix = "  │   │    "
				}

				exLine := treeLine(exPrefix+"╰─ ") + exampleText("\""+def.Example+"\"")
				if plainLen(exLine) > width {
					exLine = truncateColored(exLine, width)
				}
				b.WriteString(exLine)
				b.WriteString("\n")
			}
		}

		if !isLastPOS {
			b.WriteString("  " + treeLine("│") + "\n")
		}
	}

	return b.String()
}

func RenderError(word string, msg string) string {
	var b strings.Builder
	b.WriteString("  " + wordHeader(word) + "\n\n")
	b.WriteString("  " + errorText("✖  "+msg) + "\n")
	return b.String()
}

func plainLen(s string) int {
	l := 0
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
		} else if inEscape && r == 'm' {
			inEscape = false
		} else if !inEscape {
			l++
		}
	}
	return l
}

func truncateColored(s string, max int) string {
	if plainLen(s) <= max {
		return s
	}
	runes := []rune(s)
	visible := 0
	inEscape := false
	end := 0
	for i, r := range runes {
		if r == '\033' {
			inEscape = true
		} else if inEscape && r == 'm' {
			inEscape = false
		} else if !inEscape {
			visible++
		}
		end = i + 1
		if visible >= max-3 {
			break
		}
	}
	return string(runes[:end]) + reset + "..."
}
