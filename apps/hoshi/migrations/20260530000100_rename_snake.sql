-- Rename tables to snake_case
ALTER TABLE "Settings" RENAME TO "settings";
ALTER TABLE "Starboards" RENAME TO "starboards";
ALTER TABLE "Log" RENAME TO "star_board_logs";

-- Rename primary key constraints
ALTER TABLE "settings" RENAME CONSTRAINT "Settings_pkey" TO "settings_pkey";
ALTER TABLE "starboards" RENAME CONSTRAINT "Starboards_pkey" TO "starboards_pkey";
ALTER TABLE "star_board_logs" RENAME CONSTRAINT "Log_pkey" TO "star_board_logs_pkey";

-- Rename columns on settings
ALTER TABLE "settings" RENAME COLUMN "guildId" TO "guild_id";
ALTER TABLE "settings" RENAME COLUMN "botUpdatesChannelId" TO "bot_updates_channel_id";
ALTER TABLE "settings" RENAME COLUMN "ignoredChannelIds" TO "ignored_channel_ids";
ALTER TABLE "settings" RENAME COLUMN "createdAt" TO "created_at";
ALTER TABLE "settings" RENAME COLUMN "updatedAt" TO "updated_at";

-- Rename columns on starboards
ALTER TABLE "starboards" RENAME COLUMN "guildId" TO "guild_id";
ALTER TABLE "starboards" RENAME COLUMN "sourceEmoji" TO "source_emoji";
ALTER TABLE "starboards" RENAME COLUMN "sourceChannelId" TO "source_channel_id";
ALTER TABLE "starboards" RENAME COLUMN "targetChannelId" TO "target_channel_id";
ALTER TABLE "starboards" RENAME COLUMN "createdAt" TO "created_at";
ALTER TABLE "starboards" RENAME COLUMN "updatedAt" TO "updated_at";

-- Rename columns on star_board_logs
ALTER TABLE "star_board_logs" RENAME COLUMN "guildId" TO "guild_id";
ALTER TABLE "star_board_logs" RENAME COLUMN "channelId" TO "channel_id";
ALTER TABLE "star_board_logs" RENAME COLUMN "messageId" TO "message_id";
ALTER TABLE "star_board_logs" RENAME COLUMN "originalMessageId" TO "original_message_id";
ALTER TABLE "star_board_logs" RENAME COLUMN "createdAt" TO "created_at";
ALTER TABLE "star_board_logs" RENAME COLUMN "updatedAt" TO "updated_at";

-- Rename indexes on settings
ALTER INDEX "Settings_guildId_key" RENAME TO "settings_guild_id_key";

-- Rename indexes on star_board_logs
ALTER INDEX "Log_messageId_key" RENAME TO "star_board_logs_message_id_key";
ALTER INDEX "Log_originalMessageId_key" RENAME TO "star_board_logs_original_message_id_key";
