package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/seraj/define/api"
	"github.com/seraj/define/cache"
	"github.com/seraj/define/dict"
	"github.com/seraj/define/tui"
)

func main() {
	force := flag.Bool("f", false, "force refresh, bypass cache")
	plain := flag.Bool("plain", false, "plain text output (no TUI)")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: define [-f] [-plain] <word>")
		os.Exit(1)
	}

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

	if *plain {
		result, err := svc.Lookup(word, *force)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Print(result)
		return
	}

	p := tea.NewProgram(tui.NewAppModel(word, svc, store))
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
