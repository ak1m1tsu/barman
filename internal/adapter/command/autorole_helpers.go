package command

import "github.com/bwmarrin/discordgo"

// AutoRole button and select custom IDs — shared with the interaction handler.
const (
	AutoRoleSetButtonID    = "autorole_set"
	AutoRoleRemoveButtonID = "autorole_remove"
	AutoRoleCancelButtonID = "autorole_cancel"
	AutoRoleSelectID       = "autorole_role_select"
)

// AutoRoleButtonsRow returns the standard three-button ActionsRow used by the
// autorole management UI (Change / Remove / Cancel).
func AutoRoleButtonsRow() discordgo.ActionsRow {
	return discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Изменить",
				Style:    discordgo.PrimaryButton,
				CustomID: AutoRoleSetButtonID,
			},
			discordgo.Button{
				Label:    "Удалить",
				Style:    discordgo.DangerButton,
				CustomID: AutoRoleRemoveButtonID,
			},
			discordgo.Button{
				Label:    "Отменить",
				Style:    discordgo.SecondaryButton,
				CustomID: AutoRoleCancelButtonID,
			},
		},
	}
}
