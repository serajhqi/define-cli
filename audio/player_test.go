package audio_test

import (
	"testing"

	"github.com/seraj/define/audio"
)

func TestDetectFindsPlayer(t *testing.T) {
	p, ok := audio.Detect()
	if !ok {
		t.Log("no system player found — audio disabled (expected in CI)")
		return
	}
	if p.Command() == "" {
		t.Error("found player should have a command")
	}
}

func TestDetectPlayerCommand(t *testing.T) {
	p, ok := audio.Detect()
	if !ok {
		t.Skip("no system player found")
	}
	cmd := p.Command()
	if cmd != "paplay" && cmd != "ffplay" && cmd != "mpv" && cmd != "afplay" {
		t.Errorf("unexpected player command: %q", cmd)
	}
}

func TestNoOpPlayer(t *testing.T) {
	p, ok := audio.Detect()
	if ok {
		t.Skip("system player found, skipping no-op test")
	}
	if p != nil {
		t.Error("no-op player should be nil when no player found")
	}
}
