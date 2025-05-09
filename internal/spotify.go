package internal

import (
	"context"
	"os"
	"regexp"

	"github.com/zmb3/spotify/v2"
	auth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

var (
	redirectURI    = os.Getenv("SPOTIFY_REDIRECT_URI")
	SpotifyAuth    = auth.New(auth.WithRedirectURL(redirectURI), auth.WithScopes(auth.ScopeUserReadPrivate))
	SpotifyChannel = make(chan *spotify.Client)
	SpotifyState   = "isitstreamablebot"

	spotifyLinkRegex = regexp.MustCompile(`https://open\.spotify\.com(/intl-[a-z]{2})?/track/([a-zA-Z0-9]+)(\?si=[a-zA-Z0-9]+)?`)
)

func NewSpotifyClient() *spotify.Client {
	ctx := context.Background()

	conf := &clientcredentials.Config{
		ClientID:     os.Getenv("SPOTIFY_ID"),
		ClientSecret: os.Getenv("SPOTIFY_SECRET"),
		TokenURL:     auth.TokenURL,
	}

	httpClient := conf.Client(ctx)
	return spotify.New(httpClient)
}

// HasSpotifyLink checks if a text contains a Spotify track link
func HasSpotifyLink(text string) bool {
	return spotifyLinkRegex.MatchString(text)
}

// ExtractSpotifyLinksFromText finds a Spotify track link in a string
func ExtractSpotifyLinkFromText(text string) string {
	return spotifyLinkRegex.FindString(text)
}

// ExtractSpotifyTrackIDFromLink extracts the track ID from a Spotify track link
func ExtractSpotifyTrackIDFromLink(link string) string {
	matches := spotifyLinkRegex.FindStringSubmatch(link)
	if matches == nil {
		return ""
	}
	return matches[2]
}
