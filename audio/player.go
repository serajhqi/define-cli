package audio

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

type Player struct {
	cmd string
}

var playerCommands = []string{"paplay", "ffplay", "mpv", "afplay"}

func Detect() (*Player, bool) {
	for _, cmd := range playerCommands {
		if _, err := exec.LookPath(cmd); err == nil {
			return &Player{cmd: cmd}, true
		}
	}
	return nil, false
}

func (p *Player) Command() string {
	return p.cmd
}

func (p *Player) Play(ctx context.Context, url string) error {
	if url == "" {
		return fmt.Errorf("no audio URL")
	}

	tmpDir := os.TempDir()
	tmpFile, err := os.CreateTemp(tmpDir, "define-audio-*.mp3")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if err := downloadFile(ctx, url, tmpFile); err != nil {
		tmpFile.Close()
		return fmt.Errorf("download audio: %w", err)
	}
	tmpFile.Close()

	args := playerArgs(p.cmd, tmpPath)
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func downloadFile(ctx context.Context, url string, w io.Writer) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d downloading audio", resp.StatusCode)
	}

	_, err = io.Copy(w, resp.Body)
	return err
}

func playerArgs(cmd, file string) []string {
	switch cmd {
	case "paplay":
		return []string{"paplay", file}
	case "ffplay":
		return []string{"ffplay", "-nodisp", "-autoexit", "-loglevel", "quiet", file}
	case "mpv":
		return []string{"mpv", "--no-video", "--no-terminal", file}
	case "afplay":
		return []string{"afplay", file}
	default:
		return []string{file}
	}
}

