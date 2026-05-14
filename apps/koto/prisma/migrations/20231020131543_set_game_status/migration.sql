/*
  Warnings:

  - You are about to drop the column `finished` on the `Game` table. All the data in the column will be lost.

*/
-- CreateEnum
CREATE TYPE "GameStatus" AS ENUM ('IN_PROGRESS', 'FAILED', 'COMPLETED');

-- AlterTable
ALTER TABLE "Game" DROP COLUMN "finished",
ADD COLUMN     "status" "GameStatus" NOT NULL DEFAULT 'IN_PROGRESS';
