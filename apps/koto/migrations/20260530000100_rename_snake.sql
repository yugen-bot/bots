-- Rename tables
ALTER TABLE "Settings" RENAME TO "settings";
ALTER TABLE "Game" RENAME TO "games";
ALTER TABLE "Guess" RENAME TO "guesses";
ALTER TABLE "PlayerHints" RENAME TO "player_hints";
ALTER TABLE "PlayerStats" RENAME TO "player_stats";

-- Rename columns on settings
ALTER TABLE "settings" RENAME COLUMN "guildId" TO "guild_id";
ALTER TABLE "settings" RENAME COLUMN "botUpdatesChannelId" TO "bot_updates_channel_id";
ALTER TABLE "settings" RENAME COLUMN "channelId" TO "channel_id";
ALTER TABLE "settings" RENAME COLUMN "pingRoleId" TO "ping_role_id";
ALTER TABLE "settings" RENAME COLUMN "pingOnlyNew" TO "ping_only_new";
ALTER TABLE "settings" RENAME COLUMN "membersCanStart" TO "members_can_start";
ALTER TABLE "settings" RENAME COLUMN "enableBackToBackCooldown" TO "enable_back_to_back_cooldown";
ALTER TABLE "settings" RENAME COLUMN "backToBackCooldown" TO "back_to_back_cooldown";
ALTER TABLE "settings" RENAME COLUMN "informCooldownAfterGuess" TO "inform_cooldown_after_guess";
ALTER TABLE "settings" RENAME COLUMN "timeLimit" TO "time_limit";
ALTER TABLE "settings" RENAME COLUMN "autoStart" TO "auto_start";
ALTER TABLE "settings" RENAME COLUMN "startAfterFirstGuess" TO "start_after_first_guess";
ALTER TABLE "settings" RENAME COLUMN "maxHints" TO "max_hints";
ALTER TABLE "settings" RENAME COLUMN "hintsUsed" TO "hints_used";
ALTER TABLE "settings" RENAME COLUMN "createdAt" TO "created_at";
ALTER TABLE "settings" RENAME COLUMN "updatedAt" TO "updated_at";

-- Rename columns on games
ALTER TABLE "games" RENAME COLUMN "guildId" TO "guild_id";
ALTER TABLE "games" RENAME COLUMN "lastMessageId" TO "last_message_id";
ALTER TABLE "games" RENAME COLUMN "endingAt" TO "ending_at";
ALTER TABLE "games" RENAME COLUMN "scheduleStarted" TO "schedule_started";
ALTER TABLE "games" RENAME COLUMN "createdAt" TO "created_at";
ALTER TABLE "games" RENAME COLUMN "updatedAt" TO "updated_at";

-- Rename columns on guesses
ALTER TABLE "guesses" RENAME COLUMN "userId" TO "user_id";
ALTER TABLE "guesses" RENAME COLUMN "gameId" TO "game_id";
ALTER TABLE "guesses" RENAME COLUMN "createdAt" TO "created_at";
ALTER TABLE "guesses" RENAME COLUMN "updatedAt" TO "updated_at";

-- Rename columns on player_hints
ALTER TABLE "player_hints" RENAME COLUMN "userId" TO "user_id";
ALTER TABLE "player_hints" RENAME COLUMN "maxHints" TO "max_hints";
ALTER TABLE "player_hints" RENAME COLUMN "lastVoteTime" TO "last_vote_time";
ALTER TABLE "player_hints" RENAME COLUMN "createdAt" TO "created_at";
ALTER TABLE "player_hints" RENAME COLUMN "updatedAt" TO "updated_at";

-- Rename columns on player_stats
ALTER TABLE "player_stats" RENAME COLUMN "userId" TO "user_id";
ALTER TABLE "player_stats" RENAME COLUMN "guildId" TO "guild_id";
ALTER TABLE "player_stats" RENAME COLUMN "inGuild" TO "in_guild";
ALTER TABLE "player_stats" RENAME COLUMN "createdAt" TO "created_at";
ALTER TABLE "player_stats" RENAME COLUMN "updatedAt" TO "updated_at";

-- Rename indexes
ALTER INDEX "Settings_guildId_key" RENAME TO "settings_guild_id_key";
ALTER INDEX "Settings_guildId_channelId_idx" RENAME TO "settings_guild_id_channel_id_idx";
ALTER INDEX "Game_guildId_idx" RENAME TO "games_guild_id_idx";
ALTER INDEX "Guess_gameId_idx" RENAME TO "guesses_game_id_idx";
ALTER INDEX "PlayerHints_userId_idx" RENAME TO "player_hints_user_id_idx";
ALTER INDEX "PlayerStats_userId_guildId_idx" RENAME TO "player_stats_user_id_guild_id_idx";

-- Rename FK constraints
ALTER TABLE "games" RENAME CONSTRAINT "Game_guildId_fkey" TO "games_guild_id_fkey";
ALTER TABLE "guesses" RENAME CONSTRAINT "Guess_gameId_fkey" TO "guesses_game_id_fkey";

-- Rename PK constraints
ALTER TABLE "settings" RENAME CONSTRAINT "Settings_pkey" TO "settings_pkey";
ALTER TABLE "games" RENAME CONSTRAINT "Game_pkey" TO "games_pkey";
ALTER TABLE "guesses" RENAME CONSTRAINT "Guess_pkey" TO "guesses_pkey";
ALTER TABLE "player_hints" RENAME CONSTRAINT "PlayerHints_pkey" TO "player_hints_pkey";
ALTER TABLE "player_stats" RENAME CONSTRAINT "PlayerStats_pkey" TO "player_stats_pkey";
