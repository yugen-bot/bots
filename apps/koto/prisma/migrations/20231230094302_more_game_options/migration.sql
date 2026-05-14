-- AlterTable
UPDATE "Settings" SET frequency = frequency * 60;
ALTER TABLE "Settings" ADD COLUMN     "autoStart" BOOLEAN NOT NULL DEFAULT false,
ADD COLUMN     "timeLimit" INTEGER NOT NULL DEFAULT 60,
ALTER COLUMN "frequency" SET DEFAULT 60;
