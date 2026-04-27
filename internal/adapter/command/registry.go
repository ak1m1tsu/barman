package command

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

// Handler is a slash command interaction handler.
type Handler func(s *discordgo.Session, i *discordgo.InteractionCreate)

// Registry maps slash commands to their handlers.
type Registry struct {
	handlers map[string]Handler
	commands []*discordgo.ApplicationCommand
}

// NewRegistry returns an empty Registry ready for command registration.
func NewRegistry() *Registry {
	return &Registry{
		handlers: make(map[string]Handler),
	}
}

// Register adds a command and its handler to the registry.
func (r *Registry) Register(cmd *discordgo.ApplicationCommand, h Handler) {
	r.commands = append(r.commands, cmd)
	r.handlers[cmd.Name] = h
}

// Handle routes an InteractionCreate event to the matching handler.
func (r *Registry) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := i.ApplicationCommandData()
	h, ok := r.handlers[data.Name]
	if !ok {
		return
	}

	log := logrus.WithFields(logrus.Fields{
		"command":  commandString(data),
		"guild_id": i.GuildID,
		"notify":   true,
	})
	if i.Member != nil && i.Member.User != nil {
		log = log.WithField("user_id", i.Member.User.ID)
	}
	log.Info("command invoked")

	h(s, i)
}

// Commands returns all registered ApplicationCommand definitions.
func (r *Registry) Commands() []*discordgo.ApplicationCommand {
	return r.commands
}

// commandString builds a human-readable command string including subcommand and options.
func commandString(data discordgo.ApplicationCommandInteractionData) string {
	s := "/" + data.Name
	if len(data.Options) == 0 {
		return s
	}
	first := data.Options[0]
	if first.Type == discordgo.ApplicationCommandOptionSubCommand {
		s += " " + first.Name
		for _, opt := range first.Options {
			s += " " + opt.Name + ":" + optionValueString(opt)
		}
		return s
	}
	for _, opt := range data.Options {
		s += " " + opt.Name + ":" + optionValueString(opt)
	}
	return s
}

// optionValueString returns a string representation of an option value.
func optionValueString(opt *discordgo.ApplicationCommandInteractionDataOption) string {
	switch opt.Type {
	case discordgo.ApplicationCommandOptionString:
		return opt.StringValue()
	case discordgo.ApplicationCommandOptionInteger:
		return fmt.Sprintf("%d", opt.IntValue())
	case discordgo.ApplicationCommandOptionBoolean:
		return fmt.Sprintf("%t", opt.BoolValue())
	case discordgo.ApplicationCommandOptionUser:
		if u := opt.UserValue(nil); u != nil {
			return u.ID
		}
	case discordgo.ApplicationCommandOptionRole:
		if r := opt.RoleValue(nil, ""); r != nil {
			return r.ID
		}
	}
	return "?"
}
