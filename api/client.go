package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Phonetic struct {
	Text  string
	Audio string
}

func FirstPhoneticText(ps []Phonetic) string {
	for _, p := range ps {
		if p.Text != "" {
			return p.Text
		}
	}
	return ""
}

type Definition struct {
	Word      string
	Phonetics []Phonetic
	Meanings  []Meaning
}

type Meaning struct {
	PartOfSpeech string
	Definitions  []Def
	Synonyms     []string
	Antonyms     []string
}

type Def struct {
	Definition string
	Example    string
}

type Client struct {
	baseURL string
	http    *http.Client
}

type Option func(*Client)

func WithBaseURL(u string) Option {
	return func(c *Client) {
		c.baseURL = u
	}
}

func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) {
		c.http = h
	}
}

func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL: "https://api.dictionaryapi.dev",
		http: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

type apiResponse struct {
	Word      string        `json:"word"`
	Phonetics []apiPhonetic `json:"phonetics"`
	Meanings  []apiMeaning  `json:"meanings"`
}

type apiPhonetic struct {
	Text  string `json:"text"`
	Audio string `json:"audio"`
}

type apiMeaning struct {
	PartOfSpeech string        `json:"partOfSpeech"`
	Definitions  []apiDef      `json:"definitions"`
	Synonyms     []string      `json:"synonyms"`
	Antonyms     []string      `json:"antonyms"`
}

type apiDef struct {
	Definition string `json:"definition"`
	Example    string `json:"example"`
}

func (c *Client) Lookup(word string) (*Definition, error) {
	url := fmt.Sprintf("%s/api/v2/entries/en/%s", c.baseURL, word)
	resp, err := c.http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("%q not found in dictionary", word)
	}

	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("Rate limited. Wait and try again.")
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("dictionary API returned an error (%d)", resp.StatusCode)
	}

	var entries []apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("%q not found in dictionary", word)
	}

	entry := entries[0]
	def := &Definition{
		Word: entry.Word,
	}

	for _, p := range entry.Phonetics {
		if p.Text != "" || p.Audio != "" {
			def.Phonetics = append(def.Phonetics, Phonetic{
				Text:  p.Text,
				Audio: p.Audio,
			})
		}
	}

	for _, m := range entry.Meanings {
		meaning := Meaning{
			PartOfSpeech: m.PartOfSpeech,
			Synonyms:     m.Synonyms,
			Antonyms:     m.Antonyms,
		}
		for _, d := range m.Definitions {
			meaning.Definitions = append(meaning.Definitions, Def{
				Definition: d.Definition,
				Example:    d.Example,
			})
		}
		def.Meanings = append(def.Meanings, meaning)
	}

	return def, nil
}
