package slashcommands

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type ResetLeaderboardModule struct {
	container *di.Container
	points    *services.PointsService
	settings  *services.SettingsService
}

func GetResetLeaderboardModule(container *di.Container) *ResetLeaderboardModule {
	return &ResetLeaderboardModule{
		container: container,
		points:    container.Get(localStatic.DiPoints).(*services.PointsService),
		settings:  container.Get(sharedStatic.DiSettings).(*services.SettingsService),
	}
}

func (m *ResetLeaderboardModule) resetLeaderboard(ctx *discordgoplus.Ctx) {
	required := true

	var memberID *string
	if opt, ok := ctx.Options["member"]; ok {
		id := opt.UserValue(ctx.Session).ID
		memberID = &id
	}

	userIDValue := ""
	if memberID != nil {
		userIDValue = *memberID
	}

	err := discordgoplus.ModalRespond(ctx, &discordgo.InteractionResponseData{
		CustomID: fmt.Sprintf("RESET_LEADERBOARD/%s", userIDValue),
		Title:    "Reset Leaderboard",
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID: "confirm",
						Label:    "Type CONFIRM to reset the leaderboard",
						Style:    discordgo.TextInputShort,
						Required: &required,
					},
				},
			},
		},
	})
	if err != nil {
		utils.Logger.Error(err)
	}
}

func (m *ResetLeaderboardModule) handleModal(ctx *discordgoplus.Ctx) {
	fields := discordgoplus.ParseModalData(ctx.ModalData)
	confirm := fields["confirm"]
	if confirm != "CONFIRM" {
		ctx.Session.InteractionRespond(ctx.Interaction, &discordgo.InteractionResponse{ //nolint:errcheck
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Reset cancelled — you must type exactly `CONFIRM`.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// Parse userID from CustomID: RESET_LEADERBOARD/<userID>
	customID := ctx.Interaction.ModalSubmitData().CustomID
	parts := strings.SplitN(customID, "/", 2)
	var userID *string
	if len(parts) > 1 && parts[1] != "" {
		id := parts[1]
		userID = &id
	}

	ctx.Session.InteractionRespond(ctx.Interaction, &discordgo.InteractionResponse{ //nolint:errcheck
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	if err := m.points.ResetLeaderboard(context.Background(), ctx.Interaction.GuildID, userID); err != nil {
		utils.Logger.Warnf("reset leaderboard failed: %v", err)
		ctx.Session.FollowupMessageCreate(ctx.Interaction, true, &discordgo.WebhookParams{ //nolint:errcheck
			Content: "Failed to reset leaderboard.",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		return
	}

	msg := "Leaderboard has been reset!"
	if userID != nil {
		msg = fmt.Sprintf("Leaderboard for <@%s> has been reset!", *userID)
	}

	ctx.Session.FollowupMessageCreate(ctx.Interaction, true, &discordgo.WebhookParams{ //nolint:errcheck
		Content: msg,
		Flags:   discordgo.MessageFlagsEphemeral,
	})
}

func (m *ResetLeaderboardModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "reset-leaderboard",
			Description: "Reset the Koto leaderboard",
			Handler:     discordgoplus.HandlerFunc(m.resetLeaderboard),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "member",
					Description: "Reset only this member's stats",
					Required:    false,
				},
			},
		},
	}
}

func (m *ResetLeaderboardModule) Modals() []*discordgoplus.Modal {
	return []*discordgoplus.Modal{
		{
			CustomID: "RESET_LEADERBOARD/:userID",
			Handler:  discordgoplus.HandlerFunc(m.handleModal),
		},
	}
}
