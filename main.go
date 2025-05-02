package main

import (
	"context"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"

	"github.com/b1tray3r/isitstreamablebot/internal/discord"
	"github.com/joho/godotenv"
)

var (
	loglevel  = slog.LevelVar{}
	gitCommit string
)

func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				gitCommit = setting.Value
			}
		}
	}

	if _, err := os.Stat("/VERSION"); err == nil {
		versionData, err := os.ReadFile("/VERSION")
		if err != nil {
			panic(err)
		}
		gitCommit = strings.TrimSpace(string(versionData))
		if err != nil {
			panic(err)
		}
	}

	loglevel.Set(slog.LevelDebug)
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: &loglevel,
	})))
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load(".env")
		if err != nil {
			panic(err)
		}
	}

	config := map[string]string{
		"TWITCH_CLIENT_ID":              os.Getenv("TWITCH_CLIENT_ID"),
		"SPOTIFY_ID":                    os.Getenv("SPOTIFY_ID"),
		"SPOTIFY_SECRET":                os.Getenv("SPOTIFY_SECRET"),
		"SPOTIFY_REDIRECT_URI":          os.Getenv("SPOTIFY_REDIRECT_URI"),
		"DISCORD_BOT_TOKEN":             os.Getenv("DISCORD_BOT_TOKEN"),
		"DISCORD_WHITELIST_GUILD_IDS":   os.Getenv("DISCORD_WHITELIST_GUILD_IDS"),
		"DISCORD_WHITELIST_CHANNEL_IDS": os.Getenv("DISCORD_WHITELIST_CHANNEL_IDS"),
	}
	for key, value := range config {
		if value == "" {
			panic("Missing environment variable: " + key)
		}
	}

	slog.Info("Starting bot", "version", gitCommit)

	guildBouncer := discord.NewGuildBouncer(
		strings.Split(config["DISCORD_WHITELIST_GUILD_IDS"], ","),
	)

	channelBouncer := discord.NewChannelBouncer(
		strings.Split(config["DISCORD_WHITELIST_CHANNEL_IDS"], ","),
	)

	slog.Debug("Guild Bouncer", "guilds", guildBouncer)
	slog.Debug("Channel Bouncer", "channels", channelBouncer)

	session, err := discord.NewSession(
		config["DISCORD_BOT_TOKEN"],
		[]discord.CommandHandler{
			discord.NewSlashCommandHandler(
				[]discord.CommandHandler{
					discord.NewVersionCommandHandler(gitCommit),
					discord.NewDJSongCheckCommandHandler(config["TWITCH_CLIENT_ID"]),
				},
			),
		},
	)
	if err != nil {
		slog.Error("Error creating Discord session", "error", err)
		return
	}

	discordbot, err := discord.NewBot(
		guildBouncer,
		channelBouncer,
		session,
	)
	if err != nil {
		slog.Error("Error creating Discord bot", "error", err)
		return
	}

	slog.Info("DiscordBot running")

	<-ctx.Done()
	slog.Info("Stopping DiscordBot...")
	discordbot.Shutdown()
}
