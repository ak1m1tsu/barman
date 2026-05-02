package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	"github.com/ak1m1tsu/barman/internal/pkg/discordutil"
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
	timeout time.Duration,
) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionMessageComponent:
			handlePrefixButton(s, i, setUC, removeUC, timeout)
		case discordgo.InteractionModalSubmit:
			handlePrefixModal(s, i, setUC, timeout)
		}
	}
}

func handlePrefixButton(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	_ *guilduc.SetPrefixUseCase,
	removeUC *guilduc.RemovePrefixUseCase,
	timeout time.Duration,
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
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
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
		}); err != nil {
			log.WithError(err).Error("prefix: failed to show modal")
		}

	case prefixResetButtonID:
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := removeUC.Execute(ctx, i.GuildID); err != nil {
			log.WithError(err).Error("failed to reset prefix")
			discordutil.RespondEphemeral(s, i, "Ошибка при сбросе префикса.")
			return
		}
		log.WithField("notify", true).Info("prefix reset to default")
		discordutil.RespondEphemeral(s, i, "Префикс сброшен до глобального значения по умолчанию.")
	}
}

func handlePrefixModal(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	setUC *guilduc.SetPrefixUseCase,
	timeout time.Duration,
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
		discordutil.RespondEphemeral(s, i, "Префикс не может быть пустым.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := setUC.Execute(ctx, i.GuildID, prefix); err != nil {
		log.WithError(err).Error("failed to set prefix")
		discordutil.RespondEphemeral(s, i, "Ошибка при сохранении префикса.")
		return
	}

	log.WithFields(logrus.Fields{"prefix": prefix, "notify": true}).Info("prefix updated")
	discordutil.RespondEphemeral(s, i, fmt.Sprintf("Префикс изменён на `%s`.", prefix))
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
