package dict

import (
	"time"

	"github.com/seraj/define/api"
	"github.com/seraj/define/cache"
	"github.com/seraj/define/output"
)

type Service struct {
	api   *api.Client
	cache *cache.Store
}

func NewService(client *api.Client, store *cache.Store) *Service {
	return &Service{
		api:   client,
		cache: store,
	}
}

func (s *Service) Lookup(word string, force bool) (string, error) {
	if !force {
		if entry, ok := s.cache.Get(word); ok {
			return output.Render(entry.Data), nil
		}
	}

	def, err := s.api.Lookup(word)
	if err != nil {
		// Try cache as fallback on API error
		if entry, ok := s.cache.Get(word); ok {
			return output.Render(entry.Data), nil
		}
		return output.RenderError(word, err.Error()), nil
	}

	preview := ""
	if len(def.Meanings) > 0 && len(def.Meanings[0].Definitions) > 0 {
		preview = def.Meanings[0].Definitions[0].Definition
		if len(preview) > 70 {
			preview = preview[:67] + "..."
		}
	}

	s.cache.Set(word, &cache.CacheEntry{
		Data:       def,
		Preview:    preview,
		LookedUpAt: time.Now(),
	})

	return output.Render(def), nil
}
