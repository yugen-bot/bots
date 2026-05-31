-- Drop bot_updates_channel_id column from settings
ALTER TABLE "settings" DROP COLUMN IF EXISTS "bot_updates_channel_id";
