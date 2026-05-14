-- AddForeignKey
ALTER TABLE "Game" ADD CONSTRAINT "Game_guildId_fkey" FOREIGN KEY ("guildId") REFERENCES "Settings"("guildId") ON DELETE RESTRICT ON UPDATE CASCADE;
