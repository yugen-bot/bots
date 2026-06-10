package middlewares

import (
	"errors"
	"slices"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
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

func checkBase(e *handler.InteractionEvent) (bool, error) {
	member := e.Member()
	if member == nil {
		return false, errors.New("member not accessible")
	}
	if len(ownerIDs) > 0 && slices.Contains(ownerIDs, member.User.ID.String()) {
		return true, nil
	}
	return false, nil
}

func checkAdmin(e *handler.InteractionEvent) (bool, error) {
	base, err := checkBase(e)
	if base || err != nil {
		return base, err
	}
	perms := e.Member().Permissions
	if perms.Has(discord.PermissionAdministrator) || perms.Has(discord.PermissionManageGuild) {
		return true, nil
	}
	return false, nil
}

func checkModerator(e *handler.InteractionEvent) (bool, error) {
	admin, err := checkAdmin(e)
	if admin || err != nil {
		return admin, err
	}
	return e.Member().Permissions.Has(discord.PermissionBanMembers), nil
}

func checkResponse(e *handler.InteractionEvent, next handler.Handler, pass bool, err error) error {
	if err != nil {
		utils.Logger.Error(err)
		return e.CreateMessage(discord.MessageCreate{
			Content: "Something went wrong.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}
	if !pass {
		utils.Logger.Debugw(
			"roles: forbidden",
			"guildID", e.GuildID(),
			"userID", func() string {
				if m := e.Member(); m != nil {
					return m.User.ID.String()
				}
				return ""
			}(),
		)
		return e.CreateMessage(discord.MessageCreate{
			Content: "You don't have permission to use this command.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}
	return next(e)
}

func OwnerMiddleware(next handler.Handler) handler.Handler {
	return func(e *handler.InteractionEvent) error {
		pass, err := checkBase(e)
		return checkResponse(e, next, pass, err)
	}
}

func GuildAdminMiddleware(next handler.Handler) handler.Handler {
	return func(e *handler.InteractionEvent) error {
		pass, err := checkAdmin(e)
		return checkResponse(e, next, pass, err)
	}
}

func GuildModeratorMiddleware(next handler.Handler) handler.Handler {
	return func(e *handler.InteractionEvent) error {
		pass, err := checkModerator(e)
		return checkResponse(e, next, pass, err)
	}
}
