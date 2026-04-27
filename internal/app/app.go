package app

import (
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/ak1m1tsu/barman/internal/adapter/command"
	"github.com/ak1m1tsu/barman/internal/adapter/handler"
	sqliterepo "github.com/ak1m1tsu/barman/internal/adapter/repository/sqlite"
	"github.com/ak1m1tsu/barman/internal/infrastructure/config"
	"github.com/ak1m1tsu/barman/internal/infrastructure/database"
	"github.com/ak1m1tsu/barman/internal/infrastructure/discord"
	nekosclient "github.com/ak1m1tsu/barman/internal/infrastructure/nekos"
	otakugifsclient "github.com/ak1m1tsu/barman/internal/infrastructure/otakugifs"
	cooldownuc "github.com/ak1m1tsu/barman/internal/usecase/cooldown"
	guilduc "github.com/ak1m1tsu/barman/internal/usecase/guild"
	memberuc "github.com/ak1m1tsu/barman/internal/usecase/member"
	reactionuc "github.com/ak1m1tsu/barman/internal/usecase/reaction"
)

// App is the composition root. It wires all dependencies and manages the bot lifecycle.
type App struct {
	bot      *discord.Bot
	db       *sql.DB
	registry *command.Registry
}

// New wires all dependencies from cfg and returns a ready-to-run App.
func New(cfg *config.Config) (*App, error) {
	log := logrus.WithField("op", "app.New")
	log.Info("initializing application")

	db, err := database.Open(cfg.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("app: open database: %w", err)
	}
	log.WithField("path", cfg.Database.Path).Info("database opened")

	bot, err := discord.New(cfg.Discord.Token, cfg.Discord.AppID, cfg.Discord.GuildID)
	if err != nil {
		db.Close() //nolint:errcheck // best-effort cleanup on init failure
		return nil, fmt.Errorf("app: create discord bot: %w", err)
	}
	log.Info("discord bot created")

	// Repositories
	guildRepo := sqliterepo.NewGuildRepository(db)
	cooldownRepo := sqliterepo.NewCooldownRepository(db)
	reactionStatsRepo := sqliterepo.NewReactionStatsRepository(db)

	// Infrastructure
	roleAssigner := discord.NewRoleAssigner(bot.Session)

	// Use cases
	setAutoRole := guilduc.NewSetAutoRole(guildRepo)
	getAutoRole := guilduc.NewGetAutoRole(guildRepo)
	removeAutoRole := guilduc.NewRemoveAutoRole(guildRepo)

	setPrefix := guilduc.NewSetPrefix(guildRepo)
	getPrefix := guilduc.NewGetPrefix(guildRepo)
	removePrefix := guilduc.NewRemovePrefix(guildRepo)

	assignAutoRole := memberuc.NewAssignAutoRole(guildRepo, roleAssigner)

	nekos := nekosclient.NewClient()
	otakugifs := otakugifsclient.NewClient()
	fetchGIF := reactionuc.NewFetchGIFWithFallback(nekos, otakugifs)

	checkAndSet := cooldownuc.NewCheckAndSet(cooldownRepo)
	incrementStat := reactionuc.NewIncrementStat(reactionStatsRepo)
	getStats := reactionuc.NewGetStats(reactionStatsRepo)

	handlerTimeout := cfg.Timeouts.Handler
	if handlerTimeout == 0 {
		handlerTimeout = config.DefaultHandlerTimeout
	}
	log.WithField("handler_timeout", handlerTimeout).Info("timeouts configured")

	if url := cfg.Notifications.ErrorWebhookURL; url != "" {
		logrus.AddHook(discord.NewErrorWebhookHook(url))
		log.Info("error webhook notifications enabled")
	}
	if url := cfg.Notifications.ActivityWebhookURL; url != "" {
		logrus.AddHook(discord.NewActivityWebhookHook(url))
		log.Info("activity webhook notifications enabled")
	}

	// Commands
	registry := command.NewRegistry()
	registry.Register(command.NewPingCommand())
	registry.Register(command.NewHelpCommand())
	registry.Register(command.NewUserInfoCommand())
	registry.Register(command.NewAutoRoleCommand(getAutoRole, handlerTimeout))
	registry.Register(command.NewReactionsCommand(getStats, cfg.Discord.OwnerIDs, handlerTimeout))
	registry.Register(command.NewReactCommand(fetchGIF, checkAndSet, incrementStat, cfg.Discord.OwnerIDs, handlerTimeout))
	registry.Register(command.NewPrefixCommand(getPrefix, handlerTimeout))

	// Event handlers
	bot.Session.AddHandler(registry.Handle)
	bot.Session.AddHandler(handler.NewMemberJoinHandler(assignAutoRole, handlerTimeout))
	bot.Session.AddHandler(handler.NewPrefixInteractionHandler(setPrefix, removePrefix, handlerTimeout))
	bot.Session.AddHandler(handler.NewAutoRoleInteractionHandler(setAutoRole, getAutoRole, removeAutoRole, handlerTimeout))

	defaultPrefix := cfg.Discord.Prefix
	if defaultPrefix == "" {
		defaultPrefix = "!"
	}
	bot.Session.AddHandler(handler.NewMessageReactHandler(guildRepo, defaultPrefix, fetchGIF, checkAndSet, incrementStat, cfg.Discord.OwnerIDs, handlerTimeout))
	bot.Session.AddHandler(handler.NewReactionsInteractionHandler(getStats, handlerTimeout))

	log.Info("all dependencies wired")

	return &App{bot: bot, db: db, registry: registry}, nil
}

// Run opens the Discord gateway session and registers slash commands with Discord.
func (a *App) Run() error {
	log := logrus.WithField("op", "app.Run")
	log.Info("opening discord session")

	if err := a.bot.Session.Open(); err != nil {
		return fmt.Errorf("app: open discord session: %w", err)
	}

	for _, cmd := range a.registry.Commands() {
		if _, err := a.bot.Session.ApplicationCommandCreate(a.bot.AppID, a.bot.GuildID, cmd); err != nil {
			log.WithError(err).WithField("command", cmd.Name).Error("failed to register slash command")
		}
	}

	logrus.Info("bot is running")
	return nil
}

// Close shuts down the Discord session and the database connection.
func (a *App) Close() error {
	log := logrus.WithField("op", "app.Close")
	log.Info("closing application")

	var errs []error
	if err := a.bot.Session.Close(); err != nil {
		log.WithError(err).Error("failed to close discord session")
		errs = append(errs, fmt.Errorf("discord: %w", err))
	}
	if err := a.db.Close(); err != nil {
		log.WithError(err).Error("failed to close database")
		errs = append(errs, fmt.Errorf("database: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("app: close: %v", errs)
	}
	return nil
}
