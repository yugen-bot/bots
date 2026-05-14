-- CreateIndex
CREATE INDEX "Game_guildId_idx" ON "Game"("guildId");

-- CreateIndex
CREATE INDEX "Guess_gameId_idx" ON "Guess"("gameId");

-- CreateIndex
CREATE INDEX "PlayerStats_userId_guildId_idx" ON "PlayerStats"("userId", "guildId");

-- CreateIndex
CREATE INDEX "Settings_guildId_channelId_idx" ON "Settings"("guildId", "channelId");
