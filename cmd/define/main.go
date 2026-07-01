package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	"github.com/seraj/define/api"
	"github.com/seraj/define/audio"
	"github.com/seraj/define/cache"
	"github.com/seraj/define/dict"
	"github.com/seraj/define/output"
	"github.com/seraj/define/tui"
)

func printHelp() {
	fmt.Print(`define — Dictionary CLI

Usage:
  define [flags] [word]
  define [flags] --history

Description:
  Look up English word definitions from the Free Dictionary API.
  Without a word, or with --history, launches the interactive TUI browser.
  With a word, prints the definition and exits (one-shot mode).

Flags:
  --history    launch interactive TUI history browser
  --play       auto-play pronunciation on startup
  --plain      plain text output, no ANSI colors or TUI
  -f           force refresh, bypass cache
  -h, --help   show this help

Examples:
  define                              Launch TUI history browser
  define --history                    Launch TUI history browser
  define hello                        Look up "hello" (one-shot, colored ANSI)
  define --play hello                 Look up "hello" and play pronunciation
  define --plain hello                Look up "hello" (plain text, no colors)
  define -f hello                     Look up "hello", force fresh API call
  echo hello | define --plain hello   Plain mode (auto-detects pipe)
`)
}

func main() {
	history := flag.Bool("history", false, "launch interactive TUI history browser")
	force := flag.Bool("f", false, "force refresh, bypass cache")
	plain := flag.Bool("plain", false, "plain text output (no TUI)")
	play := flag.Bool("play", false, "auto-play pronunciation on startup")
	flag.Usage = printHelp
	flag.Parse()

	word := flag.Arg(0)

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot find cache dir:", err)
		os.Exit(1)
	}

	store, err := cache.NewStore(filepath.Join(cacheDir, "define", "cache.json"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot init cache:", err)
		os.Exit(1)
	}

	client := api.NewClient()
	svc := dict.NewService(client, store)
	player, _ := audio.Detect()

	isTerminal := term.IsTerminal(int(os.Stdout.Fd()))

	if *history {
		p := tea.NewProgram(tui.NewAppModel("", svc, store, player))
		if _, err := p.Run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	if *plain || !isTerminal {
		if word == "" {
			fmt.Fprintln(os.Stderr, "Usage: define --plain <word>")
			os.Exit(1)
		}
		defRaw, err := svc.LookupDefinition(word, *force)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Print(output.RenderPlain(defRaw))
		return
	}

	if word != "" {
		def, err := svc.LookupDefinition(word, *force)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Print(output.Render(def))

		if *play && player != nil {
			for _, p := range def.Phonetics {
				if p.Audio != "" {
					player.Play(context.Background(), word, p.Audio, *force)
					break
				}
			}
		}
		return
	}

	p := tea.NewProgram(tui.NewAppModel("", svc, store, player))
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
