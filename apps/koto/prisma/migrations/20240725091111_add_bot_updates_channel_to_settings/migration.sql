-- AlterTable
ALTER TABLE "PlayerStats" RENAME CONSTRAINT "Player_pkey" TO "PlayerStats_pkey";

-- AlterTable
ALTER TABLE "Settings" ADD COLUMN     "botUpdatesChannelId" TEXT;
