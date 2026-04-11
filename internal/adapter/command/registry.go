package command

import "github.com/bwmarrin/discordgo"

// Handler is a slash command interaction handler.
type Handler func(s *discordgo.Session, i *discordgo.InteractionCreate)

// Registry maps slash commands to their handlers.
type Registry struct {
	handlers map[string]Handler
	commands []*discordgo.ApplicationCommand
}

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
	if h, ok := r.handlers[i.ApplicationCommandData().Name]; ok {
		h(s, i)
	}
}

// Commands returns all registered ApplicationCommand definitions.
func (r *Registry) Commands() []*discordgo.ApplicationCommand {
	return r.commands
}
