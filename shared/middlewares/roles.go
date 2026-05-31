package middlewares

import (
	"errors"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

// ownerIDs is set once at startup via InitMiddlewares.
var ownerIDs []string

// InitMiddlewares stores the owner ID list from config so the package-level
// middleware functions can use it without reading os.Getenv on every request.
func InitMiddlewares(container *di.Container) {
	cfg := container.Get(static.DiConfig).(*config.Config)
	ownerIDs = cfg.OwnerIDs
}

func checkBase(ctx *discordgoplus.Ctx) (bool, error) {
	if ctx.Interaction == nil || ctx.Interaction.Member == nil {
		return false, errors.New("member not accessible")
	}

	if len(ownerIDs) > 0 &&
		slices.Contains(ownerIDs, ctx.Interaction.Member.User.ID) {
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

	if perms&discordgo.PermissionAdministrator != 0 {
		return true, nil
	}

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
			utils.Logger.Errorw(
				"roles: forbidden response failed",
				"error",
				err,
				"guildID",
				ctx.Interaction.GuildID,
				"userID",
				ctx.Interaction.Member.User.ID,
			)
		}

		return
	}

	ctx.Next()
}

func OwnerMiddleware(ctx *discordgoplus.Ctx) {
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
