package tui

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/magus-1/mfp/internal/feed"
	"github.com/magus-1/mfp/internal/player"
)

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#BD93F9")).
			MarginLeft(2)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2")).
			Background(lipgloss.Color("#44475A"))

	playingTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#50FA7B")).
				Bold(true)

	helpTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
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

// messages
type episodesLoadedMsg struct{ episodes []feed.Episode }
type episodesErrMsg struct{ err error }

type Model struct {
	list          list.Model
	player        *player.Player
	episodes      []feed.Episode
	nowPlaying    *feed.Episode
	nowPlayingIdx int
	paused        bool
	autoPlay      bool
	shuffle       bool
	err           string
	width, height int
	loading       bool
	spinner       spinner.Model
}

func NewModel() Model {
	// spinner for loading
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9"))

	episodes, _ := feed.FetchEpisodes()

	items := make([]list.Item, len(episodes))
	for i, ep := range episodes {
		items[i] = item{ep: ep}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#BD93F9")).
		BorderForeground(lipgloss.Color("#BD93F9"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#6272A4"))

	l := list.New(items, delegate, 0, 0)
	l.Title = "Music for Programming"
	l.Styles.Title = titleStyle
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)

	return Model{
		list:     l,
		player:   player.New(),
		episodes: episodes, // start empty
		autoPlay: true,
		loading:  true, // show spinner until feed loads
		spinner:  s,
	}
}

// ── Init ──────────────────────────────────────────────────────────────────────

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		fetchEpisodesCmd(),
	)
}

func fetchEpisodesCmd() tea.Cmd {
	return func() tea.Msg {
		episodes, err := feed.FetchEpisodes()
		if err != nil {
			return episodesErrMsg{err}
		}
		return episodesLoadedMsg{episodes}
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (m *Model) playNext() {
	if len(m.episodes) == 0 {
		return
	}
	var nextIdx int
	if m.shuffle {
		nextIdx = rand.Intn(len(m.episodes))
	} else {
		nextIdx = (m.nowPlayingIdx + 1) % len(m.episodes)
	}
	ep := m.episodes[nextIdx]
	m.nowPlayingIdx = nextIdx
	m.nowPlaying = &ep
	m.paused = false
	m.err = ""
	m.list.Select(nextIdx)
	if err := m.player.Play(ep.URL); err != nil {
		m.err = fmt.Sprintf("player error: %v", err)
		m.nowPlaying = nil
	}
}

func (m *Model) playSelected() {
	selected, ok := m.list.SelectedItem().(item)
	if !ok {
		return
	}
	ep := selected.ep
	// find index in episodes slice so playNext stays in sync
	for i, e := range m.episodes {
		if e.Number == ep.Number {
			m.nowPlayingIdx = i
			break
		}
	}
	m.nowPlaying = &ep
	m.paused = false
	m.err = ""
	if err := m.player.Play(ep.URL); err != nil {
		m.err = fmt.Sprintf("player error: %v", err)
		m.nowPlaying = nil
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case episodesLoadedMsg:
		m.loading = false
		m.episodes = msg.episodes
		items := make([]list.Item, len(msg.episodes))
		for i, ep := range msg.episodes {
			ep := ep
			items[i] = item{ep: ep}
		}
		m.list.SetItems(items)
		return m, nil

	case episodesErrMsg:
		m.loading = false
		m.err = fmt.Sprintf("could not load feed: %v", msg.err)
		log.Printf("feed error: %v", msg.err)
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-2)
		return m, nil

	case tea.KeyMsg:
		if m.list.FilterState() != list.Filtering {
			switch msg.String() {

			case "ctrl+c", "q":
				m.player.Stop()
				return m, tea.Quit

			case "enter":
				m.playSelected()
				return m, nil

			case " ":
				if m.player.IsPlaying() {
					m.player.Stop()
					m.paused = true
				} else if m.nowPlaying != nil && m.paused {
					m.paused = false
					_ = m.player.Play(m.nowPlaying.URL)
				}
				return m, nil

			case "n":
				m.playNext()
				return m, nil

			case "s":
				m.player.Stop()
				m.nowPlaying = nil
				m.paused = false
				return m, nil

			case "r":
				m.shuffle = !m.shuffle
				return m, nil

			case "a":
				m.autoPlay = !m.autoPlay
				return m, nil
			}
		}

		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m Model) View() string {
	if m.loading {
		return appStyle.Render(
			"\n\n  " + m.spinner.View() + "  loading episodes from musicforprogramming.net...",
		)
	}

	return appStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			m.list.View(),
			m.statusBar(),
		),
	)
}

func (m Model) statusBar() string {
	width := m.width - appStyle.GetHorizontalPadding()

	if m.err != "" {
		return errorStyle.Width(width).Render("✗ " + m.err)
	}

	// build indicators for shuffle/autoplay
	indicators := ""
	if m.shuffle {
		indicators += " 🔀"
	}
	if m.autoPlay {
		indicators += " ↻"
	}

	var left, right string

	if m.nowPlaying == nil {
		left = "■ stopped" + indicators
		right = "enter: play · space: pause · n: next · r: shuffle · a: autoplay · q: quit"
	} else {
		icon := "▶"
		state := "playing"
		if m.paused {
			icon = "⏸"
			state = "paused"
		}
		left = playingTextStyle.Render(icon+" "+state) +
			fmt.Sprintf("  %02d. %s%s", m.nowPlaying.Number, m.nowPlaying.Title, indicators)
		right = helpTextStyle.Render("space: pause · n: next · s: stop · r: shuffle · q: quit")
	}

	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	spacerWidth := width - leftWidth - rightWidth
	if spacerWidth < 1 {
		spacerWidth = 1
	}
	spacer := lipgloss.NewStyle().Width(spacerWidth).Render("")

	return statusBarStyle.
		Width(width).
		Render(left + spacer + right)
}
