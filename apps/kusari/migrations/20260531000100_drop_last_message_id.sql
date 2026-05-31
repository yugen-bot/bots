-- Drop last_message_id column from games
ALTER TABLE "games" DROP COLUMN IF EXISTS "last_message_id";
