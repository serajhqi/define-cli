package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/seraj/define/api"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *api.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	client := api.NewClient(api.WithBaseURL(srv.URL))
	return srv, client
}

func TestLookupSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/entries/en/hello" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"word":"hello","phonetics":[{"text":"/həˈləʊ/"},{"text":"/həˈloʊ/"}],"meanings":[{"partOfSpeech":"noun","definitions":[{"definition":"\"Hello!\" or an equivalent greeting.","example":""}],"synonyms":["greeting"]},{"partOfSpeech":"verb","definitions":[{"definition":"To greet with \"hello\".","example":""}]},{"partOfSpeech":"interjection","definitions":[{"definition":"A greeting (salutation) said when meeting someone.","example":"Hello, everyone."},{"definition":"A greeting used when answering the telephone.","example":"Hello? How may I help you?"}]}]}]`))
	}))
	defer srv.Close()

	client := api.NewClient(api.WithBaseURL(srv.URL))
	def, err := client.Lookup("hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if def.Word != "hello" {
		t.Errorf("word = %q, want %q", def.Word, "hello")
	}
	if len(def.Phonetics) != 2 {
		t.Fatalf("phonetics count = %d, want 2", len(def.Phonetics))
	}
	if def.Phonetics[0].Text != "/həˈləʊ/" {
		t.Errorf("phonetics[0] = %q, want %q", def.Phonetics[0].Text, "/həˈləʊ/")
	}
	if def.Phonetics[1].Text != "/həˈloʊ/" {
		t.Errorf("phonetics[1] = %q, want %q", def.Phonetics[1].Text, "/həˈloʊ/")
	}
	if len(def.Meanings) != 3 {
		t.Fatalf("meanings count = %d, want 3", len(def.Meanings))
	}

	m0 := def.Meanings[0]
	if m0.PartOfSpeech != "noun" {
		t.Errorf("meanings[0].PartOfSpeech = %q, want %q", m0.PartOfSpeech, "noun")
	}
	if len(m0.Definitions) != 1 {
		t.Fatalf("meanings[0] defs = %d, want 1", len(m0.Definitions))
	}
	if m0.Definitions[0].Definition != `"Hello!" or an equivalent greeting.` {
		t.Errorf("noun def = %q", m0.Definitions[0].Definition)
	}

	m2 := def.Meanings[2]
	if m2.PartOfSpeech != "interjection" {
		t.Errorf("meanings[2].PartOfSpeech = %q, want %q", m2.PartOfSpeech, "interjection")
	}
	if len(m2.Definitions) != 2 {
		t.Fatalf("interjection defs = %d, want 2", len(m2.Definitions))
	}
	if m2.Definitions[0].Example != "Hello, everyone." {
		t.Errorf("interjection example = %q", m2.Definitions[0].Example)
	}
}

func TestLookupNotFound(t *testing.T) {
	srv, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer srv.Close()

	_, err := client.Lookup("zxcvbnm")
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

func TestLookupServerError(t *testing.T) {
	srv, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer srv.Close()

	_, err := client.Lookup("hello")
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
	if !strings.Contains(err.Error(), "API returned an error") {
		t.Errorf("error should mention API error, got: %v", err)
	}
}

func TestLookupTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer srv.Close()

	client := api.NewClient(
		api.WithBaseURL(srv.URL),
		api.WithHTTPClient(&http.Client{Timeout: 10 * time.Millisecond}),
	)

	_, err := client.Lookup("hello")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "network error") {
		t.Errorf("error should mention network, got: %v", err)
	}
}

func TestLookupRateLimited(t *testing.T) {
	srv, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})
	defer srv.Close()

	_, err := client.Lookup("hello")
	if err == nil {
		t.Fatal("expected error for 429, got nil")
	}
	if !strings.Contains(err.Error(), "Rate limited") {
		t.Errorf("error should mention rate limit, got: %v", err)
	}
}
