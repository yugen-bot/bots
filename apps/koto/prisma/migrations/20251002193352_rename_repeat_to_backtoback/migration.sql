/*
  Warnings:

  - You are about to drop the column `enableRepeatCooldown` on the `Settings` table. All the data in the column will be lost.
  - You are about to drop the column `repeatCooldown` on the `Settings` table. All the data in the column will be lost.

*/
-- RenameColum
ALTER TABLE "Settings" RENAME COLUMN "enableRepeatCooldown" TO "enableBackToBackCooldown";
ALTER TABLE "Settings" RENAME COLUMN "repeatCooldown" TO "backToBackCooldown";
