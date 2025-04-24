package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
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
	loglevel.Set(slog.LevelDebug)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	listeningChannelIDs := strings.Split(config["DISCORD_LISTENING_CHANNEL_IDS"], ",")

	client := internal.NewSpotifyClient()

	discord, err := discordgo.New("Bot " + config["DISCORD_BOT_TOKEN"])
	if err != nil {
		log.Fatalf("error creating Discord session: %v", err)
	}

	slog.Info("Starting Discord bot...")
	_ = discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

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
					slog.Error("Spotify error", "err", err, "trackID", spotifyID)
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

					messages := []string{"Twitch DJ Catalog Results for \"" + track.Name + "\":"}
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
