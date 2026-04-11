package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

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
	configPath := flag.String("config", "configs/config.yaml", "Path to YAML config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := database.Open(cfg.Database.Path)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	bot, err := discord.New(cfg.Discord.Token, cfg.Discord.AppID, cfg.Discord.GuildID)
	if err != nil {
		log.Fatalf("discord: %v", err)
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
		log.Fatalf("discord: open session: %v", err)
	}
	defer bot.Session.Close()

	// Register slash commands with Discord
	for _, cmd := range registry.Commands() {
		if _, err := bot.Session.ApplicationCommandCreate(bot.AppID, bot.GuildID, cmd); err != nil {
			log.Printf("discord: register command %q: %v", cmd.Name, err)
		}
	}

	log.Println("Bot is running. Press Ctrl+C to stop.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down...")
}
