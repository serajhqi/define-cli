package output

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/seraj/define/api"
)

const width = 70

const (
	reset        = "\033[0m"
	bold         = "\033[1m"
	dim          = "\033[38;5;245m"
	white        = "\033[37m"
	brightYellow = "\033[93m"
	cyan         = "\033[36m"
	blue         = "\033[94m"
	red          = "\033[91m"
)

func wordHeader(s string) string     { return bold + brightYellow + s + reset }
func phoneticLabel(s string) string   { return blue + s + reset }
func posLabel(s string) string       { return bold + cyan + s + reset }
func definitionText(s string) string { return white + s + reset }
func exampleText(s string) string    { return dim + s + reset }
func treeLine(s string) string       { return dim + s + reset }
func errorText(s string) string      { return red + s + reset }

func visibleLen(s string) int {
	n := 0
	inEsc := false
	for _, r := range s {
		if r == '\033' {
			inEsc = true
		} else if inEsc {
			if r == 'm' {
				inEsc = false
			}
		} else {
			n++
		}
	}
	return n
}

func wrapText(text string, maxWidth int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	var cur []string
	curLen := 0

	for _, w := range words {
		wlen := utf8.RuneCountInString(w)
		if len(cur) == 0 {
			cur = append(cur, w)
			curLen = wlen
		} else if curLen+1+wlen <= maxWidth {
			cur = append(cur, w)
			curLen += 1 + wlen
		} else {
			lines = append(lines, strings.Join(cur, " "))
			cur = []string{w}
			curLen = wlen
		}
	}
	if len(cur) > 0 {
		lines = append(lines, strings.Join(cur, " "))
	}
	return lines
}

func Render(def *api.Definition) string {
	var b strings.Builder

	b.WriteString("\n  " + wordHeader(def.Word))
	phonText := api.FirstPhoneticText(def.Phonetics)
	if phonText != "" {
		b.WriteString("  " + phoneticLabel(phonText))
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
			isLastDef := di == len(meaning.Definitions)-1

			var prefix string
			if isLastPOS {
				prefix = "      "
			} else {
				prefix = "  │   "
			}

			connector := "├─ "
			if isLastDef {
				connector = "╰─ "
			}

			num := fmt.Sprintf("%d. ", di+1)
			treePart := treeLine(prefix + connector)
			contentStart := visibleLen(treePart) + len(num)

			var contPart string
			var contPad int
			if isLastDef {
				contPart = treeLine(prefix)
				contPad = contentStart - visibleLen(contPart)
			} else {
				contPart = treeLine(prefix + "│")
				contPad = contentStart - visibleLen(contPart)
			}

			wrapped := wrapText(def.Definition, width-contentStart)
			for li, w := range wrapped {
				if li == 0 {
					b.WriteString(treePart + definitionText(num+w))
				} else {
					b.WriteString(contPart + strings.Repeat(" ", contPad) + definitionText(w))
				}
				b.WriteString("\n")
			}

			if def.Example != "" {
				var exPrefix string
				switch {
				case !isLastPOS && !isLastDef:
					exPrefix = "  │   │    "
				case !isLastPOS && isLastDef:
					exPrefix = "  │        "
				case isLastPOS && !isLastDef:
					exPrefix = "      │    "
				default:
					exPrefix = "           "
				}

				exTree := treeLine(exPrefix + "╰─ ")
				exContentStart := visibleLen(exTree) + 1 // +1 for opening quote
				exContPart := treeLine(exPrefix)
				exContPad := exContentStart - visibleLen(exContPart)

				wrappedEx := wrapText(def.Example, width-exContentStart-1) // -1 for closing quote
				for li, w := range wrappedEx {
					switch {
					case len(wrappedEx) == 1:
						b.WriteString(exTree + exampleText("\""+w+"\""))
					case li == 0:
						b.WriteString(exTree + exampleText("\""+w))
					case li == len(wrappedEx)-1:
						b.WriteString(exContPart + strings.Repeat(" ", exContPad) + exampleText(w+"\""))
					default:
						b.WriteString(exContPart + strings.Repeat(" ", exContPad) + exampleText(w))
					}
					b.WriteString("\n")
				}
			}

			if len(meaning.Definitions) > 1 && !isLastDef {
				b.WriteString(treeLine(prefix+"│") + "\n")
			}
		}

		if !isLastPOS {
			b.WriteString("  " + treeLine("│") + "\n")
		}
	}

	b.WriteString("\n")
	return b.String()
}

func RenderError(word string, msg string) string {
	var b strings.Builder
	b.WriteString("  " + wordHeader(word) + "\n\n")
	b.WriteString("  " + errorText("✖  "+msg) + "\n")
	return b.String()
}

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func RenderPlain(def *api.Definition) string {
	colored := Render(def)
	return ansiPattern.ReplaceAllString(colored, "")
}

func RenderErrorPlain(word, msg string) string {
	colored := RenderError(word, msg)
	return ansiPattern.ReplaceAllString(colored, "")
}
