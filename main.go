package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/b1tray3r/isitstreamablebot/internal"
	"github.com/b1tray3r/isitstreamablebot/internal/twitch"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/zmb3/spotify/v2"
)

var (
	loglevel = slog.LevelVar{}
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load environment variables from a .env file if it exists
	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
	}

	config := map[string]string{
		"TWITCH_CLIENT_ID":              os.Getenv("TWITCH_CLIENT_ID"),
		"SPOTIFY_ID":                    os.Getenv("SPOTIFY_ID"),
		"SPOTIFY_SECRET":                os.Getenv("SPOTIFY_SECRET"),
		"SPOTIFY_REDIRECT_URI":          os.Getenv("SPOTIFY_REDIRECT_URI"),
		"DISCORD_BOT_TOKEN":             os.Getenv("DISCORD_BOT_TOKEN"),
		"DISCORD_LISTENING_CHANNEL_IDS": os.Getenv("DISCORD_LISTENING_CHANNEL_IDS"),
	}
	for key, value := range config {
		if value == "" {
			log.Fatalf("Missing environment variable: %s", key)
		}
	}

	// I want the loglevel to be configurable dynamically via HTTP request - but initially set to info
	loglevel.Set(slog.LevelInfo)
	slogHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: &loglevel})
	slog.SetDefault(slog.New(slogHandler))

	// first start an HTTP server
	http.HandleFunc("/callback", internal.CompleteAuth)
	http.HandleFunc("POST /loglevel/{level}", func(w http.ResponseWriter, r *http.Request) {
		level := r.PathValue("level")
		if level == "" {
			http.Error(w, "Missing log level in URL path", http.StatusBadRequest)
			return
		}

		var newLevel slog.Level
		switch strings.ToLower(level) {
		case "debug":
			newLevel = slog.LevelDebug
		case "info":
			newLevel = slog.LevelInfo
		case "warn", "warning":
			newLevel = slog.LevelWarn
		case "error":
			newLevel = slog.LevelError
		default:
			http.Error(w, "Invalid log level", http.StatusBadRequest)
			return
		}

		loglevel.Set(newLevel)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Log level set to %s", newLevel.String())
	})
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	listeningChannelIDs := strings.Split(config["DISCORD_LISTENING_CHANNEL_IDS"], ",")

	url := internal.SpotifyAuth.AuthURL(internal.SpotifyState)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)
	// wait for auth to complete
	client := <-internal.SpotifyChannel

	discord, err := discordgo.New("Bot " + config["DISCORD_BOT_TOKEN"])
	if err != nil {
		log.Fatalf("error creating Discord session: %v", err)
	}

	slog.Info("Starting Discord bot...")
	_ = discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		// Check if the message is posted in a channel we listen on
		isListeningChannel := false
		for _, channelID := range listeningChannelIDs {
			if m.ChannelID == strings.TrimSpace(channelID) {
				isListeningChannel = true
				break
			}
		}

		if !isListeningChannel {
			return
		}

		if internal.HasSpotifyLink(m.Content) {
			links := internal.ExtractSpotifyLinksFromText(m.Content)
			for _, link := range links {
				spotifyID := internal.ExtractSpotifyTrackIDFromLink(link)
				slog.Info("Extracted", "SpotifyID", spotifyID)
				s.MessageReactionAdd(m.ChannelID, m.ID, "⏳")

				trackID := spotify.ID(spotifyID)
				track, err := client.GetTrack(ctx, trackID)
				if err != nil {
					slog.Error("Spotify error", "err", err)
					return
				}
				if track == nil {
					slog.Error("Track not found on Spotify")
					return
				}

				response, err := twitch.SendRequest(track.Name)
				if err != nil {
					slog.Error("TwitchDJCatalog error", "err", err)
					return
				}

				for _, r := range response {
					if len(r.Data.SearchDJCatalog.Edges) == 0 {
						slog.Error("No results found.")
						return
					}

					messages := []string{"Twich DJ Catalog Results for \"" + track.Name + "\":"}
					for _, edge := range r.Data.SearchDJCatalog.Edges {
						state := "✅"
						if edge.Node.IsBlockedTrack {
							state = "❌"
						}

						artists := []string{}
						for _, artist := range edge.Node.Artists {
							artists = append(artists, artist.Name)
						}
						sort.Strings(artists)

						indent := ""
						if track.Name == edge.Node.Title {
							spotifyDuration := int(track.Duration / 1000)
							nodeDuration := int(edge.Node.Duration)
							if spotifyDuration == nodeDuration {
								// At this point we know that this is the requested track - so we
								// remove the loading reaction and add the state reaction
								// and indent the message in the list
								s.MessageReactionRemove(m.ChannelID, m.ID, "⏳", s.State.User.ID)
								s.MessageReactionAdd(m.ChannelID, m.ID, state)
								indent = ">   "
							}
						}

						durationMinutes := int(edge.Node.Duration) / 60
						durationSeconds := int(edge.Node.Duration) % 60
						messages = append(messages, fmt.Sprintf(
							"%s%s: %s - %s - %s - %dm %ds",
							indent,
							state,
							edge.Node.Title,
							strings.Join(artists, ", "),
							strings.Join(edge.Node.Genres, ", "),
							durationMinutes,
							durationSeconds,
						))
					}

					s.ChannelMessageSendReply(m.ChannelID, strings.Join(messages, "\n"), m.Reference())
				}
			}
		}
	})
	err = discord.Open()
	if err != nil {
		log.Fatalf("error opening connection: %v", err)
	}

	<-ctx.Done()
	slog.Info("Stopping Discord bot...")
	discord.Close()
}
