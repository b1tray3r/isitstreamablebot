package main

import (
	"context"
	"database/sql"
	"embed"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/b1tray3r/isitstreamablebot/internal"
	"github.com/b1tray3r/isitstreamablebot/internal/discord"
	"github.com/b1tray3r/isitstreamablebot/internal/discord/handler"
	commandhandler "github.com/b1tray3r/isitstreamablebot/internal/discord/handler/command"
	"github.com/b1tray3r/isitstreamablebot/internal/discord/handler/interaction"
	"github.com/b1tray3r/isitstreamablebot/internal/discord/handler/message"
	"github.com/b1tray3r/isitstreamablebot/internal/store"
	"github.com/b1tray3r/isitstreamablebot/internal/watchlist"
	"github.com/joho/godotenv"
	migrate "github.com/rubenv/sql-migrate"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed sql/schemas/*
var dbMigrations embed.FS

var (
	loglevel  = slog.LevelVar{}
	gitCommit string
)

// init initializes the logger and reads the version from the build info or VERSION file.
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

	slog.Info("Preparing database")
	db, err := sql.Open("sqlite3", "./data/bot.db?mode=memory&cache=shared")
	if err != nil {
		slog.Error("Error opening database", "error", err)
		return
	}
	migrations := migrate.EmbedFileSystemMigrationSource{
		FileSystem: dbMigrations,
		Root:       "sql/schemas",
	}
	n, err := migrate.Exec(db, "sqlite3", migrations, migrate.Up)
	if err != nil {
		slog.Error("Error applying migrations", "error", err)
		return
	}
	slog.Debug("Applied migrations", "count", n)

	store := store.NewStore(db)

	slog.Info("Starting bot", "version", gitCommit)
	guildBouncer := discord.NewGuildBouncer(
		strings.Split(config["DISCORD_WHITELIST_GUILD_IDS"], ","),
	)

	channelBouncer := discord.NewChannelBouncer(
		strings.Split(config["DISCORD_WHITELIST_CHANNEL_IDS"], ","),
	)

	bouncerFunc := func(guildID, channelID string) bool {
		if !guildBouncer.Check(guildID) {
			slog.Info("Guild bouncer check failed", "guildID", guildID)
			return false
		}
		if !channelBouncer.Check(channelID) {
			slog.Info("Channel bouncer check failed", "channelID", channelID)
			return false
		}
		return true
	}

	wl := watchlist.NewWatchlist(store)

	session, err := discord.NewSession(
		config["DISCORD_BOT_TOKEN"],
		[]handler.CommandHandler{
			commandhandler.NewSlashCommandHandler(
				[]handler.CommandHandler{
					commandhandler.NewVersionCommandHandler(gitCommit),
					commandhandler.NewDJSongCheckCommandHandler(config["TWITCH_CLIENT_ID"], bouncerFunc),
					commandhandler.NewListWatchlistCommandHandler(wl),
				},
			),
		},
		[]handler.InteractionHandler{
			interaction.NewButtonHandler(
				[]handler.InteractionHandler{
					interaction.NewWatchlistHandler(
						wl,
						discord.WATCHLIST_ADD,
					),
				},
			),
		},
		[]handler.MessageHandler{
			message.NewSpotifyCheckHandler(internal.NewSpotifyClient(), store),
		},
	)
	if err != nil {
		slog.Error("Error creating Discord session", "error", err)
		return
	}

	slog.Info("DiscordBot running")

	slog.Info("Starting background song checker")
	checker := watchlist.NewChecker(store, 2*time.Second)
	checker.StartWatchlistChecker(ctx)

	<-ctx.Done()
	slog.Info("Stopping DiscordBot...")
	session.Close()
}
