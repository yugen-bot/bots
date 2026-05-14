/*
  Warnings:

  - You are about to drop the column `metaData` on the `Game` table. All the data in the column will be lost.

*/
-- AlterTable
ALTER TABLE "Game" DROP COLUMN "metaData";

-- AlterTable
ALTER TABLE "Guess" ADD COLUMN     "meta" JSONB NOT NULL DEFAULT '{}';
