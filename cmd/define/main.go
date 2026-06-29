package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/seraj/define/api"
	"github.com/seraj/define/cache"
	"github.com/seraj/define/dict"
)

func main() {
	force := flag.Bool("f", false, "force refresh, bypass cache")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: define [-f] <word>")
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

	result, err := svc.Lookup(word, *force)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Print(result)
}
