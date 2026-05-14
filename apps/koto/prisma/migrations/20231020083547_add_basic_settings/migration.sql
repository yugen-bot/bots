-- AlterTable
ALTER TABLE "Settings" ADD COLUMN     "channelId" TEXT,
ADD COLUMN     "cooldown" INTEGER NOT NULL DEFAULT 10,
ADD COLUMN     "frequency" INTEGER NOT NULL DEFAULT 1,
ADD COLUMN     "pingRoleId" TEXT;
