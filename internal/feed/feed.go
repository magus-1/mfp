package feed

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

const feedURL = "https://musicforprogramming.net/rss.php"

type Episode struct {
	Number int
	Title  string
	URL    string
}

// FetchEpisodes fetches live episodes from musicforprogramming.net.
// Returns a descriptive error if anything goes wrong.
func FetchEpisodes() ([]Episode, error) {
	log.Printf("fetching feed from %s", feedURL)

	client := &http.Client{Timeout: 10 * time.Second}
	fp := gofeed.NewParser()
	fp.Client = client

	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch feed from %s: %w", feedURL, err)
	}

	log.Printf("feed fetched OK: %q — %d items", feed.Title, len(feed.Items))

	if len(feed.Items) == 0 {
		return nil, fmt.Errorf("feed was empty — check %s manually", feedURL)
	}

	var episodes []Episode
	for _, item := range feed.Items {
		if len(item.Enclosures) == 0 {
			log.Printf("skipping %q — no enclosure (no audio URL)", item.Title)
			continue
		}

		ep := Episode{
			Title: cleanTitle(item.Title),
			URL:   item.Enclosures[0].URL,
		}

		// Title format is "N: Artist" — extract number if present
		ep.Number = parseNumber(item.Title)

		log.Printf("  episode %02d: %s (%s)", ep.Number, ep.Title, ep.URL)
		episodes = append(episodes, ep)
	}

	if len(episodes) == 0 {
		return nil, fmt.Errorf("feed had %d items but none had audio enclosures", len(feed.Items))
	}

	log.Printf("loaded %d episodes", len(episodes))
	return episodes, nil
}

// cleanTitle strips the episode number prefix, e.g. "1: Datassette" -> "Datassette"
func cleanTitle(raw string) string {
	if idx := strings.Index(raw, ": "); idx != -1 {
		return strings.TrimSpace(raw[idx+2:])
	}
	return strings.TrimSpace(raw)
}

// parseNumber extracts the leading number from "N: Title", returns 0 if none found.
func parseNumber(raw string) int {
	if idx := strings.Index(raw, ":"); idx != -1 {
		n, err := strconv.Atoi(strings.TrimSpace(raw[:idx]))
		if err == nil {
			return n
		}
	}
	return 0
}

// HardcodedEpisodes is kept as a fallback for offline dev/testing.
func HardcodedEpisodes() []Episode {
	return []Episode{
		{Number: 1, Title: "datassette", URL: "https://datasette.io/mfp/mfp1.mp3"},
		{Number: 2, Title: "cc licensed beats", URL: "https://datasette.io/mfp/mfp2.mp3"},
		{Number: 3, Title: "four tet", URL: "https://datasette.io/mfp/mfp3.mp3"},
	}
}
