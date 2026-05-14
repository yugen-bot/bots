-- AlterTable
ALTER TABLE "Settings" ALTER COLUMN "cooldown" SET DEFAULT 600,
ALTER COLUMN "backToBackCooldown" SET DEFAULT 600;

UPDATE "Settings" SET "cooldown" = "cooldown" * 60, "backToBackCooldown" = "cooldown" * 60;
