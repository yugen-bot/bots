-- Rename tables
ALTER TABLE "Settings" RENAME TO "settings";
ALTER TABLE "Game" RENAME TO "games";
ALTER TABLE "History" RENAME TO "histories";
ALTER TABLE "PlayerStats" RENAME TO "player_stats";
ALTER TABLE "PlayerSaves" RENAME TO "player_saves";

-- Rename columns on settings
ALTER TABLE "settings" RENAME COLUMN "guildId" TO "guild_id";
ALTER TABLE "settings" RENAME COLUMN "botUpdatesChannelId" TO "bot_updates_channel_id";
ALTER TABLE "settings" RENAME COLUMN "channelId" TO "channel_id";
ALTER TABLE "settings" RENAME COLUMN "shameRoleId" TO "shame_role_id";
ALTER TABLE "settings" RENAME COLUMN "removeShameRoleAfterHighscore" TO "remove_shame_role_after_highscore";
ALTER TABLE "settings" RENAME COLUMN "lastShameUserId" TO "last_shame_user_id";
ALTER TABLE "settings" RENAME COLUMN "highscoreDate" TO "highscore_date";
ALTER TABLE "settings" RENAME COLUMN "maxSaves" TO "max_saves";
ALTER TABLE "settings" RENAME COLUMN "savesUsed" TO "saves_used";
ALTER TABLE "settings" RENAME COLUMN "createdAt" TO "created_at";
ALTER TABLE "settings" RENAME COLUMN "updatedAt" TO "updated_at";

-- Rename columns on games
ALTER TABLE "games" RENAME COLUMN "guildId" TO "guild_id";
ALTER TABLE "games" RENAME COLUMN "lastMessageId" TO "last_message_id";
ALTER TABLE "games" RENAME COLUMN "isHighscored" TO "is_highscored";
ALTER TABLE "games" RENAME COLUMN "createdAt" TO "created_at";
ALTER TABLE "games" RENAME COLUMN "updatedAt" TO "updated_at";

-- Rename columns on histories
ALTER TABLE "histories" RENAME COLUMN "userId" TO "user_id";
ALTER TABLE "histories" RENAME COLUMN "messageId" TO "message_id";
ALTER TABLE "histories" RENAME COLUMN "gameId" TO "game_id";
ALTER TABLE "histories" RENAME COLUMN "createdAt" TO "created_at";
ALTER TABLE "histories" RENAME COLUMN "updatedAt" TO "updated_at";

-- Rename columns on player_stats
ALTER TABLE "player_stats" RENAME COLUMN "userId" TO "user_id";
ALTER TABLE "player_stats" RENAME COLUMN "guildId" TO "guild_id";
ALTER TABLE "player_stats" RENAME COLUMN "inGuild" TO "in_guild";
ALTER TABLE "player_stats" RENAME COLUMN "createdAt" TO "created_at";
ALTER TABLE "player_stats" RENAME COLUMN "updatedAt" TO "updated_at";

-- Rename columns on player_saves
ALTER TABLE "player_saves" RENAME COLUMN "userId" TO "user_id";
ALTER TABLE "player_saves" RENAME COLUMN "maxSaves" TO "max_saves";
ALTER TABLE "player_saves" RENAME COLUMN "lastVoteTime" TO "last_vote_time";
ALTER TABLE "player_saves" RENAME COLUMN "createdAt" TO "created_at";
ALTER TABLE "player_saves" RENAME COLUMN "updatedAt" TO "updated_at";

-- Rename indexes
ALTER INDEX "Settings_guildId_key" RENAME TO "settings_guild_id_key";
ALTER INDEX "Settings_guildId_idx" RENAME TO "settings_guild_id_idx";
ALTER INDEX "Game_guildId_idx" RENAME TO "games_guild_id_idx";
ALTER INDEX "History_userId_gameId_idx" RENAME TO "histories_user_id_game_id_idx";
ALTER INDEX "PlayerStats_userId_guildId_idx" RENAME TO "player_stats_user_id_guild_id_idx";
ALTER INDEX "PlayerSaves_userId_idx" RENAME TO "player_saves_user_id_idx";

-- Rename FK constraints
ALTER TABLE "games" RENAME CONSTRAINT "Game_guildId_fkey" TO "games_guild_id_fkey";
ALTER TABLE "histories" RENAME CONSTRAINT "History_gameId_fkey" TO "histories_game_id_fkey";

-- Rename PK constraints
ALTER TABLE "settings" RENAME CONSTRAINT "Settings_pkey" TO "settings_pkey";
ALTER TABLE "games" RENAME CONSTRAINT "Game_pkey" TO "games_pkey";
ALTER TABLE "histories" RENAME CONSTRAINT "History_pkey" TO "histories_pkey";
ALTER TABLE "player_stats" RENAME CONSTRAINT "PlayerStats_pkey" TO "player_stats_pkey";
ALTER TABLE "player_saves" RENAME CONSTRAINT "PlayerSaves_pkey" TO "player_saves_pkey";
