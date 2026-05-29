-- History(game_id) - helps cascade queries
CREATE INDEX "histories_game_id_idx" ON "histories" ("game_id");

-- PlayerStats(guild_id) - guild leaderboard scans
CREATE INDEX "player_stats_guild_id_idx" ON "player_stats" ("guild_id");
