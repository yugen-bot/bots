package middlewares

import (
	"errors"
	"slices"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

var ownerIDs []string

func InitMiddlewares(container *di.Container) {
	cfg := container.Get(static.DiConfig).(*config.Config)
	ownerIDs = cfg.OwnerIDs
}

func checkBase(ctx *disgoplus.Ctx) (bool, error) {
	if ctx.Member == nil {
		return false, errors.New("member not accessible")
	}

	if len(ownerIDs) > 0 &&
		slices.Contains(ownerIDs, ctx.Member.User.ID.String()) {
		return true, nil
	}

	return false, nil
}

func checkAdmin(ctx *disgoplus.Ctx) (bool, error) {
	base, err := checkBase(ctx)
	if base || err != nil {
		return base, err
	}

	perms := ctx.Member.Permissions

	if perms.Has(discord.PermissionAdministrator) {
		return true, nil
	}

	if perms.Has(discord.PermissionManageGuild) {
		return true, nil
	}

	return false, nil
}

func checkModerator(ctx *disgoplus.Ctx) (bool, error) {
	admin, err := checkAdmin(ctx)
	if admin || err != nil {
		return admin, err
	}

	return ctx.Member.Permissions.Has(discord.PermissionBanMembers), nil
}

func checkResponse(ctx *disgoplus.Ctx, pass bool, err error) {
	if err != nil {
		utils.Logger.Error(err)

		resErr := disgoplus.ErrorResponse(ctx)
		if resErr != nil {
			utils.Logger.Error(resErr)
		}

		return
	}

	if !pass {
		err := disgoplus.ForbiddenResponse(ctx)
		if err != nil {
			utils.Logger.Errorw(
				"roles: forbidden response failed",
				"error",
				err,
				"guildID",
				ctx.GuildID,
				"userID",
				ctx.Member.User.ID,
			)
		}

		return
	}

	ctx.Next()
}

func OwnerMiddleware(ctx *disgoplus.Ctx) {
	pass, err := checkBase(ctx)
	checkResponse(ctx, pass, err)
}

func GuildAdminMiddleware(ctx *disgoplus.Ctx) {
	pass, err := checkAdmin(ctx)
	checkResponse(ctx, pass, err)
}

func GuildModeratorMiddleware(ctx *disgoplus.Ctx) {
	pass, err := checkModerator(ctx)
	checkResponse(ctx, pass, err)
}
