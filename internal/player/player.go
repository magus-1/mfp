package player

import (
    "os/exec"
)

// Player holds the running mpv process so we can stop it.
type Player struct {
    cmd *exec.Cmd
}

// Play starts streaming the given URL via mpv.
// It kills any previously running stream first.
func (p *Player) Play(url string) error {
    p.Stop()

    // --really-quiet suppresses mpv output in the terminal
    // --no-video ensures we never try to render video
    p.cmd = exec.Command("mpv", "--really-quiet", "--no-video", url)

    return p.cmd.Start()
}

// Stop kills the current mpv process if one is running.
func (p *Player) Stop() {
    if p.cmd != nil && p.cmd.Process != nil {
        _ = p.cmd.Process.Kill()
        p.cmd = nil
    }
}

// IsPlaying returns true if a stream is currently active.
func (p *Player) IsPlaying() bool {
    return p.cmd != nil && p.cmd.Process != nil
}

// New returns a fresh Player.
func New() *Player {
    return &Player{}
}
