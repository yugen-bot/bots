-- DropForeignKey
ALTER TABLE "Game" DROP CONSTRAINT "Game_guildId_fkey";

-- DropForeignKey
ALTER TABLE "Guess" DROP CONSTRAINT "Guess_gameId_fkey";

-- AddForeignKey
ALTER TABLE "Game" ADD CONSTRAINT "Game_guildId_fkey" FOREIGN KEY ("guildId") REFERENCES "Settings"("guildId") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Guess" ADD CONSTRAINT "Guess_gameId_fkey" FOREIGN KEY ("gameId") REFERENCES "Game"("id") ON DELETE CASCADE ON UPDATE CASCADE;
