package app

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
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
	bot          *discord.Bot
	db           *sql.DB
	registry     *command.Registry
	webhookHooks []*discord.WebhookHook
	activity     config.ActivityConfig
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
		if cerr := db.Close(); cerr != nil {
			logrus.WithError(cerr).Warn("app: failed to close database during init cleanup")
		}
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

	var webhookHooks []*discord.WebhookHook
	if url := cfg.Notifications.WebhookURL; url != "" {
		errHook := discord.NewErrorWebhookHook(url)
		actHook := discord.NewActivityWebhookHook(url)
		logrus.AddHook(errHook)
		logrus.AddHook(actHook)
		webhookHooks = append(webhookHooks, errHook, actHook)
		log.Info("webhook notifications enabled")
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

	return &App{
		bot:          bot,
		db:           db,
		registry:     registry,
		webhookHooks: webhookHooks,
		activity:     cfg.Discord.Activity,
	}, nil
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

	if a.activity.Text != "" {
		if err := a.applyActivity(); err != nil {
			log.WithError(err).Warn("failed to set bot activity")
		}
	}

	logrus.Info("bot is running")
	return nil
}

// applyActivity sets the bot's gateway presence. Only name, type, and state are
// rendered by Discord clients for bot presence; other Rich Presence fields are ignored.
func (a *App) applyActivity() error {
	act := &discordgo.Activity{
		Name:  a.activity.Text,
		State: a.activity.State,
	}

	switch a.activity.Type {
	case "watching":
		act.Type = discordgo.ActivityTypeWatching
	case "listening":
		act.Type = discordgo.ActivityTypeListening
	case "competing":
		act.Type = discordgo.ActivityTypeCompeting
	default: // "playing" or empty
		act.Type = discordgo.ActivityTypeGame
	}

	return a.bot.Session.UpdateStatusComplex(discordgo.UpdateStatusData{
		Status:     "online",
		Activities: []*discordgo.Activity{act},
	})
}

// Close drains in-flight webhook goroutines, then shuts down the Discord
// session and the database connection. Webhook drain is given 5 seconds; if
// that expires, shutdown continues anyway and the context error is logged.
func (a *App) Close() error {
	log := logrus.WithField("op", "app.Close")
	log.Info("closing application")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, h := range a.webhookHooks {
		if err := h.Shutdown(shutdownCtx); err != nil {
			log.WithError(err).Warn("webhook hook: shutdown timed out, some notifications may be lost")
		}
	}

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
