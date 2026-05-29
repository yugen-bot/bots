-- Create enum types
CREATE TYPE "GameStatus" AS ENUM ('IN_PROGRESS', 'FAILED', 'COMPLETED');
CREATE TYPE "GameType" AS ENUM ('NORMAL');

-- Create Settings table
CREATE TABLE "Settings" (
    "id"                            SERIAL NOT NULL,
    "guildId"                       TEXT NOT NULL,
    "botUpdatesChannelId"           TEXT,
    "channelId"                     TEXT,
    "cooldown"                      INTEGER NOT NULL DEFAULT 0,
    "math"                          BOOLEAN NOT NULL DEFAULT true,
    "shameRoleId"                   TEXT,
    "removeShameRoleAfterHighscore" BOOLEAN NOT NULL DEFAULT false,
    "lastShameUserId"               TEXT,
    "highscore"                     INTEGER NOT NULL DEFAULT 0,
    "highscoreDate"                 TIMESTAMP(3),
    "saves"                         DOUBLE PRECISION NOT NULL DEFAULT 0,
    "maxSaves"                      DOUBLE PRECISION NOT NULL DEFAULT 2,
    "savesUsed"                     DOUBLE PRECISION NOT NULL DEFAULT 0,
    "createdAt"                     TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt"                     TIMESTAMP(3) NOT NULL,
    CONSTRAINT "Settings_pkey" PRIMARY KEY ("id")
);

-- Create Game table
CREATE TABLE "Game" (
    "id"            SERIAL NOT NULL,
    "guildId"       TEXT NOT NULL,
    "lastMessageId" TEXT,
    "status"        "GameStatus" NOT NULL DEFAULT 'IN_PROGRESS',
    "type"          "GameType" NOT NULL DEFAULT 'NORMAL',
    "isHighscored"  BOOLEAN NOT NULL DEFAULT false,
    "createdAt"     TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt"     TIMESTAMP(3) NOT NULL,
    CONSTRAINT "Game_pkey" PRIMARY KEY ("id")
);

-- Create History table
CREATE TABLE "History" (
    "id"        SERIAL NOT NULL,
    "userId"    TEXT NOT NULL,
    "messageId" TEXT,
    "gameId"    INTEGER NOT NULL,
    "number"    INTEGER NOT NULL,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    CONSTRAINT "History_pkey" PRIMARY KEY ("id")
);

-- Create PlayerStats table
CREATE TABLE "PlayerStats" (
    "id"        SERIAL NOT NULL,
    "userId"    TEXT NOT NULL,
    "guildId"   TEXT NOT NULL,
    "inGuild"   BOOLEAN NOT NULL DEFAULT true,
    "points"    INTEGER NOT NULL DEFAULT 0,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    CONSTRAINT "PlayerStats_pkey" PRIMARY KEY ("id")
);

-- Create PlayerSaves table
CREATE TABLE "PlayerSaves" (
    "id"            SERIAL NOT NULL,
    "userId"        TEXT NOT NULL,
    "saves"         DOUBLE PRECISION NOT NULL DEFAULT 0,
    "maxSaves"      DOUBLE PRECISION NOT NULL DEFAULT 2,
    "lastVoteTime"  TIMESTAMP(3),
    "createdAt"     TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt"     TIMESTAMP(3) NOT NULL,
    CONSTRAINT "PlayerSaves_pkey" PRIMARY KEY ("id")
);

-- Unique and regular indexes for Settings
CREATE UNIQUE INDEX "Settings_guildId_key" ON "Settings"("guildId");
CREATE INDEX "Settings_guildId_idx" ON "Settings"("guildId");

-- Indexes for Game
CREATE INDEX "Game_guildId_idx" ON "Game"("guildId");

-- Indexes for History
CREATE INDEX "History_userId_gameId_idx" ON "History"("userId", "gameId");

-- Indexes for PlayerStats
CREATE INDEX "PlayerStats_userId_guildId_idx" ON "PlayerStats"("userId", "guildId");

-- Indexes for PlayerSaves
CREATE INDEX "PlayerSaves_userId_idx" ON "PlayerSaves"("userId");

-- Foreign keys
ALTER TABLE "Game"
    ADD CONSTRAINT "Game_guildId_fkey"
    FOREIGN KEY ("guildId") REFERENCES "Settings"("guildId")
    ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "History"
    ADD CONSTRAINT "History_gameId_fkey"
    FOREIGN KEY ("gameId") REFERENCES "Game"("id")
    ON DELETE CASCADE ON UPDATE CASCADE;
