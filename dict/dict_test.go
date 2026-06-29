package dict_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/seraj/define/api"
	"github.com/seraj/define/cache"
	"github.com/seraj/define/dict"
)

func TestLookupAPIAndCache(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"word":"hello","phonetics":[{"text":"/həˈloʊ/"}],"meanings":[{"partOfSpeech":"interjection","definitions":[{"definition":"A greeting.","example":"Hello, world!"}]}]}]`))
	}))
	defer srv.Close()

	dir := t.TempDir()
	cacheStore, err := cache.NewStore(filepath.Join(dir, "cache.json"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	client := api.NewClient(api.WithBaseURL(srv.URL))
	svc := dict.NewService(client, cacheStore)

	// First lookup: API call
	result, err := svc.Lookup("hello", false)
	if err != nil {
		t.Fatalf("first lookup: %v", err)
	}
	if callCount != 1 {
		t.Errorf("first lookup API calls = %d, want 1", callCount)
	}
	if !strings.Contains(result, "hello") {
		t.Error("result should contain word")
	}
	if !strings.Contains(result, "/həˈloʊ/") {
		t.Error("result should contain phonetic")
	}

	// Second lookup: should use cache
	result2, err := svc.Lookup("hello", false)
	if err != nil {
		t.Fatalf("second lookup: %v", err)
	}
	if callCount != 1 {
		t.Errorf("second lookup API calls = %d, want 1 (from cache)", callCount)
	}
	if result2 != result {
		t.Error("cached result should match first result")
	}
}

func TestLookupForceRefresh(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"word":"hello","phonetics":[{"text":"/həˈloʊ/"}],"meanings":[{"partOfSpeech":"interjection","definitions":[{"definition":"A greeting.","example":"Hello, world!"}]}]}]`))
	}))
	defer srv.Close()

	dir := t.TempDir()
	cacheStore, _ := cache.NewStore(filepath.Join(dir, "cache.json"))
	client := api.NewClient(api.WithBaseURL(srv.URL))
	svc := dict.NewService(client, cacheStore)

	// First lookup via API
	_, err := svc.Lookup("hello", false)
	if err != nil {
		t.Fatalf("first lookup: %v", err)
	}
	if callCount != 1 {
		t.Errorf("first lookup API calls = %d, want 1", callCount)
	}

	// Force refresh
	_, err = svc.Lookup("hello", true)
	if err != nil {
		t.Fatalf("force lookup: %v", err)
	}
	if callCount != 2 {
		t.Errorf("force lookup API calls = %d, want 2", callCount)
	}
}

func TestLookupNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	dir := t.TempDir()
	cacheStore, _ := cache.NewStore(filepath.Join(dir, "cache.json"))
	client := api.NewClient(api.WithBaseURL(srv.URL))
	svc := dict.NewService(client, cacheStore)

	result, err := svc.Lookup("zxcvbnm", false)
	if err != nil {
		t.Fatalf("should not return error for not-found: %v", err)
	}
	if !strings.Contains(result, "zxcvbnm") {
		t.Error("should contain word")
	}
	if !strings.Contains(result, "not found") {
		t.Error("should contain error message")
	}
	if !strings.Contains(result, "✖") {
		t.Error("should contain error marker")
	}
}
