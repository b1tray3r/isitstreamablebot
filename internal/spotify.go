package internal

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"regexp"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

var redirectURI = os.Getenv("SPOTIFY_REDIRECT_URI")
var (
	SpotifyAuth    = spotifyauth.New(spotifyauth.WithRedirectURL(redirectURI), spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate))
	SpotifyChannel = make(chan *spotify.Client)
	SpotifyState   = "isitstreamablebot"
)

// https://open.spotify.com/track/0WVgnZcUUvQVoaa2gEnv3Q?si=5e37c0a62f894f84
var spotifyLinkRegex = regexp.MustCompile(`https://open\.spotify\.com(/intl-de)?/track/([a-zA-Z0-9]+)(\?si=[a-zA-Z0-9]+)?`)

func HasSpotifyLink(text string) bool {
	return spotifyLinkRegex.MatchString(text)
}

func ExtractSpotifyLinksFromText(text string) []string {
	matches := spotifyLinkRegex.FindAllString(text, -1)
	if matches == nil {
		return nil
	}
	return matches
}
func ExtractSpotifyTrackIDFromLink(link string) string {
	matches := spotifyLinkRegex.FindStringSubmatch(link)
	if matches == nil {
		return ""
	}
	return matches[2]
}

func CompleteAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := SpotifyAuth.Token(r.Context(), SpotifyState, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatalf("Couldn't get token: %v", err)
	}
	if st := r.FormValue("state"); st != SpotifyState {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, SpotifyState)
	}

	client := spotify.New(SpotifyAuth.Client(r.Context(), tok))
	slog.Info("Spotify client created", "client", client)
	SpotifyChannel <- client
}
