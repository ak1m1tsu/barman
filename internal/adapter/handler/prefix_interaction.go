package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	guilduc "github.com/ak1m1tsu/barman/internal/usecase/guild"
)

const (
	prefixSetButtonID   = "prefix_set"
	prefixResetButtonID = "prefix_reset"
	prefixModalID       = "prefix_modal"
	prefixInputID       = "prefix_value"
)

// NewPrefixInteractionHandler handles button clicks and modal submits for /prefix.
func NewPrefixInteractionHandler(
	setUC *guilduc.SetPrefixUseCase,
	removeUC *guilduc.RemovePrefixUseCase,
) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionMessageComponent:
			handlePrefixButton(s, i, setUC, removeUC)
		case discordgo.InteractionModalSubmit:
			handlePrefixModal(s, i, setUC)
		}
	}
}

func handlePrefixButton(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	_ *guilduc.SetPrefixUseCase,
	removeUC *guilduc.RemovePrefixUseCase,
) {
	data := i.MessageComponentData()
	log := logrus.WithFields(logrus.Fields{
		"guild_id":  i.GuildID,
		"button_id": data.CustomID,
	})
	if i.Member != nil && i.Member.User != nil {
		log = log.WithField("user_id", i.Member.User.ID)
	}

	switch data.CustomID {
	case prefixSetButtonID:
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{ //nolint:errcheck
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: prefixModalID,
				Title:    "Изменить префикс",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID:    prefixInputID,
								Label:       "Новый префикс",
								Style:       discordgo.TextInputShort,
								Placeholder: "!",
								Required:    true,
								MinLength:   1,
								MaxLength:   5,
							},
						},
					},
				},
			},
		})

	case prefixResetButtonID:
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := removeUC.Execute(ctx, i.GuildID); err != nil {
			log.WithError(err).Error("failed to reset prefix")
			respondComponentEphemeral(s, i, "Ошибка при сбросе префикса.")
			return
		}
		respondComponentEphemeral(s, i, "Префикс сброшен до глобального значения по умолчанию.")
	}
}

func handlePrefixModal(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	setUC *guilduc.SetPrefixUseCase,
) {
	data := i.ModalSubmitData()
	if data.CustomID != prefixModalID {
		return
	}

	log := logrus.WithField("guild_id", i.GuildID)
	if i.Member != nil && i.Member.User != nil {
		log = log.WithField("user_id", i.Member.User.ID)
	}

	prefix := modalTextValue(data.Components, prefixInputID)
	if prefix == "" {
		respondComponentEphemeral(s, i, "Префикс не может быть пустым.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := setUC.Execute(ctx, i.GuildID, prefix); err != nil {
		log.WithError(err).Error("failed to set prefix")
		respondComponentEphemeral(s, i, "Ошибка при сохранении префикса.")
		return
	}

	respondComponentEphemeral(s, i, fmt.Sprintf("Префикс изменён на `%s`.", prefix))
}

// modalTextValue extracts the submitted value for a TextInput with the given customID.
func modalTextValue(components []discordgo.MessageComponent, customID string) string {
	for _, c := range components {
		row, ok := c.(*discordgo.ActionsRow)
		if !ok {
			continue
		}
		for _, inner := range row.Components {
			ti, ok := inner.(*discordgo.TextInput)
			if ok && ti.CustomID == customID {
				return ti.Value
			}
		}
	}
	return ""
}

func respondComponentEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{ //nolint:errcheck
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
