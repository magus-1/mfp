package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/magus-1/mfp/internal/feed"
	"github.com/magus-1/mfp/internal/player"
)

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#C9A0DC")).
			MarginLeft(2)

	// Status bar base — full width blue background
	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#6C91BF"))

	// These render TEXT ONLY — no background — the outer statusBarStyle provides it
	playingTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A8CC8C")).
				Bold(true)

	helpTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Faint(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E88388")).
			Bold(true)

	appStyle = lipgloss.NewStyle().Padding(1, 2)
)

// ── List item adapter ─────────────────────────────────────────────────────────

type item struct {
	ep feed.Episode
}

func (i item) Title() string       { return fmt.Sprintf("%02d. %s", i.ep.Number, i.ep.Title) }
func (i item) Description() string { return i.ep.URL }
func (i item) FilterValue() string { return i.ep.Title }

// ── Model ─────────────────────────────────────────────────────────────────────

type Model struct {
	list          list.Model
	player        *player.Player
	nowPlaying    *feed.Episode
	paused        bool
	err           string
	width, height int
}

func NewModel() Model {
	episodes, _ := feed.FetchEpisodes()

	items := make([]list.Item, len(episodes))
	for i, ep := range episodes {
		ep := ep
		items[i] = item{ep: ep}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#C9A0DC")).
		BorderForeground(lipgloss.Color("#C9A0DC"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#9B7FBE"))

	l := list.New(items, delegate, 0, 0)
	l.Title = "Music for Programming"
	l.Styles.Title = titleStyle
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)

	return Model{
		list:   l,
		player: player.New(),
	}
}

// ── Init ──────────────────────────────────────────────────────────────────────

func (m Model) Init() tea.Cmd {
	return nil
}

// ── Update ────────────────────────────────────────────────────────────────────

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-2)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.player.Stop()
			return m, tea.Quit

		case "enter":
			if m.list.FilterState() == list.Filtering {
				break
			}
			selected, ok := m.list.SelectedItem().(item)
			if !ok {
				break
			}
			ep := selected.ep
			m.nowPlaying = &ep
			m.paused = false
			m.err = ""
			if err := m.player.Play(ep.URL); err != nil {
				m.nowPlaying = nil
				m.err = fmt.Sprintf("player error: %v", err)
			}
			return m, nil

		case "p":
			if m.list.FilterState() == list.Filtering {
				break
			}
			if m.player.IsPlaying() {
				m.player.Stop()
				m.paused = true
			} else if m.nowPlaying != nil && m.paused {
				m.paused = false
				_ = m.player.Play(m.nowPlaying.URL)
			}
			return m, nil

		case "s":
			if m.list.FilterState() == list.Filtering {
				break
			}
			m.player.Stop()
			m.nowPlaying = nil
			m.paused = false
			return m, nil
		}

		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m Model) View() string {
	return appStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			m.list.View(),
			m.statusBar(),
		),
	)
}

func (m Model) statusBar() string {
	// Total usable width (appStyle has 2 padding on each side)
	width := m.width - appStyle.GetHorizontalPadding()

	if m.err != "" {
		return errorStyle.
			Width(width).
			Render("✗ " + m.err)
	}

	var left, right string

	if m.nowPlaying == nil {
		left = "■ stopped"
		right = "enter: play · p: pause · s: stop · q: quit"
	} else {
		icon := "▶"
		state := "playing"
		if m.paused {
			icon = "⏸"
			state = "paused"
		}
		left = playingTextStyle.Render(icon+" "+state) +
			fmt.Sprintf("  %02d. %s", m.nowPlaying.Number, m.nowPlaying.Title)
		right = helpTextStyle.Render("p: pause · s: stop · q: quit")
	}

	// Render left and right as plain strings first to measure them
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)

	// Spacer fills remaining space so right is flush to the edge
	spacerWidth := width - leftWidth - rightWidth
	if spacerWidth < 1 {
		spacerWidth = 1
	}
	spacer := lipgloss.NewStyle().Width(spacerWidth).Render("")

	// Apply the blue background ONCE across the full composed string
	return statusBarStyle.
		Width(width).
		Render(left + spacer + right)
}
