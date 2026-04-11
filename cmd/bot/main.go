package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/ak1m1tsu/barman/internal/adapter/command"
	"github.com/ak1m1tsu/barman/internal/adapter/handler"
	sqliterepo "github.com/ak1m1tsu/barman/internal/adapter/repository/sqlite"
	"github.com/ak1m1tsu/barman/internal/infrastructure/config"
	"github.com/ak1m1tsu/barman/internal/infrastructure/database"
	"github.com/ak1m1tsu/barman/internal/infrastructure/discord"
	guilduc "github.com/ak1m1tsu/barman/internal/usecase/guild"
	memberuc "github.com/ak1m1tsu/barman/internal/usecase/member"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)

	configPath := flag.String("config", "configs/config.yaml", "Path to YAML config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		logrus.WithError(err).Fatal("config: failed to load")
	}

	db, err := database.Open(cfg.Database.Path)
	if err != nil {
		logrus.WithError(err).Fatal("database: failed to open")
	}
	defer func() {
		if err := db.Close(); err != nil {
			logrus.WithError(err).Error("database: failed to close")
		}
	}()

	bot, err := discord.New(cfg.Discord.Token, cfg.Discord.AppID, cfg.Discord.GuildID)
	if err != nil {
		logrus.WithError(err).Fatal("discord: failed to create bot")
	}

	// Wire dependencies
	guildRepo := sqliterepo.NewGuildRepository(db)
	roleAssigner := discord.NewRoleAssigner(bot.Session)

	setAutoRole := guilduc.NewSetAutoRole(guildRepo)
	getAutoRole := guilduc.NewGetAutoRole(guildRepo)
	removeAutoRole := guilduc.NewRemoveAutoRole(guildRepo)
	assignAutoRole := memberuc.NewAssignAutoRole(guildRepo, roleAssigner)

	// Register commands
	registry := command.NewRegistry()
	registry.Register(command.NewPingCommand())
	registry.Register(command.NewHelpCommand())
	registry.Register(command.NewUserInfoCommand())
	registry.Register(command.NewAutoRoleCommand(setAutoRole, getAutoRole, removeAutoRole))

	bot.Session.AddHandler(registry.Handle)
	bot.Session.AddHandler(handler.NewMemberJoinHandler(assignAutoRole))

	if err := bot.Session.Open(); err != nil {
		logrus.WithError(err).Fatal("discord: failed to open session")
	}
	defer func() {
		if err := bot.Session.Close(); err != nil {
			logrus.WithError(err).Error("discord: failed to close session")
		}
	}()

	// Register slash commands with Discord
	for _, cmd := range registry.Commands() {
		if _, err := bot.Session.ApplicationCommandCreate(bot.AppID, bot.GuildID, cmd); err != nil {
			logrus.WithError(err).WithField("command", cmd.Name).Error("discord: failed to register command")
		}
	}

	logrus.Info("bot is running")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logrus.Info("shutting down")
}
