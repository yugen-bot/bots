// Package utils contains internal helpers for the hachimitsu bot.
package utils

import "github.com/disgoorg/disgo/discord"

// IsPrivileged returns true when a member should be exempt from honeypot
// enforcement. Exempt permissions: Administrator, Manage Server, Manage
// Messages, Ban Members, Kick Members.
func IsPrivileged(perms discord.Permissions) bool {
	return perms.Has(discord.PermissionAdministrator) ||
		perms.Has(discord.PermissionManageGuild) ||
		perms.Has(discord.PermissionManageMessages) ||
		perms.Has(discord.PermissionBanMembers) ||
		perms.Has(discord.PermissionKickMembers)
}
