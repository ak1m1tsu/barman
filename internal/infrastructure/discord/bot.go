package discord

import "github.com/bwmarrin/discordgo"

// Bot wraps a discordgo session with application metadata.
type Bot struct {
	Session *discordgo.Session
	AppID   string
	GuildID string
}

// New creates a Bot with the required gateway intents for guild member events.
func New(token, appID, guildID string) (*Bot, error) {
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	s.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsMessageContent

	return &Bot{
		Session: s,
		AppID:   appID,
		GuildID: guildID,
	}, nil
}
