-- Convert ignored_channel_ids from text[] to jsonb, preserving existing data

-- Add new jsonb column with empty array default
ALTER TABLE "settings" ADD COLUMN "ignored_channel_ids_new" jsonb NOT NULL DEFAULT '[]'::jsonb;

-- Copy existing data: to_jsonb converts text[] to a JSON array
UPDATE "settings" SET "ignored_channel_ids_new" = to_jsonb("ignored_channel_ids");

-- Drop old text[] column
ALTER TABLE "settings" DROP COLUMN "ignored_channel_ids";

-- Rename new column into place
ALTER TABLE "settings" RENAME COLUMN "ignored_channel_ids_new" TO "ignored_channel_ids";
