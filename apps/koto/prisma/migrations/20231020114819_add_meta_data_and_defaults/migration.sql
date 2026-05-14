-- AlterTable
ALTER TABLE "Game" ADD COLUMN     "metaData" JSONB NOT NULL DEFAULT '{}';

-- AlterTable
ALTER TABLE "Guess" ALTER COLUMN "points" SET DEFAULT 0;
