package feed

// Episode represents a single musicforprogramming.net episode.
type Episode struct {
	Number int
	Title  string
	URL    string
}

// Hardcoded episodes for the milestone. Replace with RSS fetch later.
func HardcodedEpisodes() []Episode {
	return []Episode{
		{
			Number: 1,
			Title:  "datassette",
			URL:    "https://datashat.net/music_for_programming_78-datassette.mp3",
		},
		{
			Number: 2,
			Title:  "placeholder",
			URL:    "https://datasette.s3.amazonaws.com/mfp/mfp2.mp3",
		},
		{
			Number: 3,
			Title:  "placeholder",
			URL:    "https://datasette.s3.amazonaws.com/mfp/mfp3.mp3",
		},
		{
			Number: 4,
			Title:  "placeholder",
			URL:    "https://datasette.s3.amazonaws.com/mfp/mfp4.mp3",
		},
		{
			Number: 5,
			Title:  "placeholder",
			URL:    "https://datasette.s3.amazonaws.com/mfp/mfp5.mp3",
		},
	}
}
