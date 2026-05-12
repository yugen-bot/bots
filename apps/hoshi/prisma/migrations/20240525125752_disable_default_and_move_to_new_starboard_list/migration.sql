/*
  Warnings:

  - You are about to drop the column `channelId` on the `Settings` table. All the data in the column will be lost.
  - You are about to drop the column `emoji` on the `Settings` table. All the data in the column will be lost.
  - You are about to drop the column `channelId` on the `SpecificChannels` table. All the data in the column will be lost.
  - Added the required column `targetChannelId` to the `SpecificChannels` table without a default value. This is not possible if the table is not empty.

*/
-- AlterTable
ALTER TABLE "SpecificChannels" RENAME COLUMN "channelId" to "targetChannelId";
ALTER TABLE "SpecificChannels" ADD COLUMN "sourceEmoji" TEXT NOT NULL DEFAULT '⭐';
ALTER TABLE "SpecificChannels" ALTER COLUMN "sourceChannelId" DROP NOT NULL;

-- Create default starboards
INSERT INTO "SpecificChannels" ("guildId", "targetChannelId", "sourceEmoji", "updatedAt", "createdAt")
SELECT "guildId", "channelId", "emoji", "updatedAt", "createdAt"
FROM "Settings"
WHERE "channelId" IS NOT NULL;


-- AlterTable
ALTER TABLE "Settings" DROP COLUMN "channelId";
ALTER TABLE "Settings" DROP COLUMN "emoji";
