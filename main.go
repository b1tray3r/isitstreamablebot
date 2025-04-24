package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/b1tray3r/isitstreamablebot/internal"
	"github.com/joho/godotenv"
)

var (
	loglevel = slog.LevelVar{}
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	loglevel.Set(slog.LevelDebug)

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

	discordbot, err := internal.NewDiscordBot(config["DISCORD_BOT_TOKEN"], listeningChannelIDs)
	if err != nil {
		log.Fatalf("error creating Discord bot: %v", err)
	}
	discordbot.SpotifyClient = client

	slog.Info("Starting Discord bot...")

	<-ctx.Done()
	slog.Info("Stopping Discord bot...")
	discordbot.Session.Close()
}
