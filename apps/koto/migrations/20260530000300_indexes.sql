-- Game(status, ending_at) - cleanup queries filter by both
CREATE INDEX "games_status_ending_at_idx" ON "games" ("status", "ending_at");

-- Guess(game_id, user_id) - leaderboard groupings
CREATE INDEX "guesses_game_id_user_id_idx" ON "guesses" ("game_id", "user_id");

-- PlayerStats(guild_id) - guild leaderboard
CREATE INDEX "player_stats_guild_id_idx" ON "player_stats" ("guild_id");
