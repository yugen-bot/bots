package middlewares

import (
	"errors"
	"os"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func checkBase(ctx *discordgoplus.Ctx) (bool, error) {
	if ctx.Interaction == nil || ctx.Interaction.Member == nil {
		return false, errors.New("Member not accessible")
	}

	interaction := ctx.Interaction

	if interaction == nil {
		return true, nil
	}

	member := interaction.Member
	if member == nil {
		return false, nil
	}

	ownerIds := strings.Split(os.Getenv(static.EnvOwnerIDs), ",")
	if len(ownerIds) > 0 && slices.Contains(ownerIds, interaction.Member.User.ID) {
		return true, nil
	}

	return false, nil
}

func checkAdmin(ctx *discordgoplus.Ctx) (bool, error) {
	base, err := checkBase(ctx)
	if base || err != nil {
		return base, err
	}

	perms := ctx.Interaction.Member.Permissions

	// admin check
	if perms&discordgo.PermissionAdministrator != 0 {
		return true, nil
	}

	// guild manage check
	if perms&discordgo.PermissionManageGuild != 0 {
		return true, nil
	}

	return false, nil
}

func checkModerator(ctx *discordgoplus.Ctx) (bool, error) {
	admin, err := checkAdmin(ctx)
	if admin || err != nil {
		return admin, err
	}

	// moderator check
	perms := ctx.Interaction.Member.Permissions
	return perms&discordgo.PermissionBanMembers != 0, nil
}

func checkResponse(ctx *discordgoplus.Ctx, pass bool, err error) {
	if err != nil {
		utils.Logger.Error(err)

		resErr := discordgoplus.ErrorResponse(ctx)
		if resErr != nil {
			utils.Logger.Error(err)
		}

		return
	}

	if !pass {
		err := discordgoplus.ForbiddenResponse(ctx)
		if err != nil {
			utils.Logger.Error(err)
		}
		return
	}

	ctx.Next()
}

func GuildOwnerMiddleware(ctx *discordgoplus.Ctx) {
	pass, err := checkBase(ctx)
	checkResponse(ctx, pass, err)
}

func GuildAdminMiddleware(ctx *discordgoplus.Ctx) {
	pass, err := checkAdmin(ctx)
	checkResponse(ctx, pass, err)
}

func GuildModeratorMiddleware(ctx *discordgoplus.Ctx) {
	pass, err := checkModerator(ctx)
	checkResponse(ctx, pass, err)
}
