-- Create enum types
CREATE TYPE "GameStatus" AS ENUM ('IN_PROGRESS', 'FAILED', 'COMPLETED', 'OUT_OF_TIME');

-- Create Settings table
CREATE TABLE "Settings" (
    "id"                        SERIAL NOT NULL,
    "guildId"                   TEXT NOT NULL,
    "botUpdatesChannelId"       TEXT,
    "channelId"                 TEXT,
    "pingRoleId"                TEXT,
    "pingOnlyNew"               BOOLEAN NOT NULL DEFAULT true,
    "membersCanStart"           BOOLEAN NOT NULL DEFAULT false,
    "cooldown"                  INTEGER NOT NULL DEFAULT 600,
    "enableBackToBackCooldown"  BOOLEAN NOT NULL DEFAULT false,
    "backToBackCooldown"        INTEGER NOT NULL DEFAULT 600,
    "informCooldownAfterGuess"  BOOLEAN NOT NULL DEFAULT false,
    "frequency"                 INTEGER NOT NULL DEFAULT 60,
    "timeLimit"                 INTEGER NOT NULL DEFAULT 60,
    "autoStart"                 BOOLEAN NOT NULL DEFAULT false,
    "startAfterFirstGuess"      BOOLEAN NOT NULL DEFAULT false,
    "hints"                     DOUBLE PRECISION NOT NULL DEFAULT 0,
    "maxHints"                  DOUBLE PRECISION NOT NULL DEFAULT 5,
    "hintsUsed"                 DOUBLE PRECISION NOT NULL DEFAULT 0,
    "createdAt"                 TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt"                 TIMESTAMP(3) NOT NULL,
    CONSTRAINT "Settings_pkey" PRIMARY KEY ("id")
);

-- Create Game table
CREATE TABLE "Game" (
    "id"              SERIAL NOT NULL,
    "guildId"         TEXT NOT NULL,
    "lastMessageId"   TEXT,
    "word"            TEXT NOT NULL,
    "endingAt"        TIMESTAMP(3) NOT NULL,
    "number"          INTEGER NOT NULL DEFAULT 1,
    "scheduleStarted" BOOLEAN NOT NULL DEFAULT true,
    "status"          "GameStatus" NOT NULL DEFAULT 'IN_PROGRESS',
    "meta"            JSONB NOT NULL DEFAULT '{}',
    "createdAt"       TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt"       TIMESTAMP(3) NOT NULL,
    CONSTRAINT "Game_pkey" PRIMARY KEY ("id")
);

-- Create Guess table
CREATE TABLE "Guess" (
    "id"        SERIAL NOT NULL,
    "userId"    TEXT NOT NULL,
    "gameId"    INTEGER NOT NULL,
    "word"      TEXT NOT NULL,
    "points"    INTEGER NOT NULL DEFAULT 0,
    "meta"      JSONB NOT NULL DEFAULT '{}',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    CONSTRAINT "Guess_pkey" PRIMARY KEY ("id")
);

-- Create PlayerHints table
CREATE TABLE "PlayerHints" (
    "id"           SERIAL NOT NULL,
    "userId"       TEXT NOT NULL,
    "hints"        DOUBLE PRECISION NOT NULL DEFAULT 0,
    "maxHints"     DOUBLE PRECISION NOT NULL DEFAULT 5,
    "lastVoteTime" TIMESTAMP(3),
    "createdAt"    TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt"    TIMESTAMP(3) NOT NULL,
    CONSTRAINT "PlayerHints_pkey" PRIMARY KEY ("id")
);

-- Create PlayerStats table
CREATE TABLE "PlayerStats" (
    "id"           SERIAL NOT NULL,
    "userId"       TEXT NOT NULL,
    "guildId"      TEXT NOT NULL,
    "inGuild"      BOOLEAN NOT NULL DEFAULT true,
    "points"       INTEGER NOT NULL DEFAULT 0,
    "participated" INTEGER NOT NULL DEFAULT 0,
    "wins"         INTEGER NOT NULL DEFAULT 0,
    "createdAt"    TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt"    TIMESTAMP(3) NOT NULL,
    CONSTRAINT "PlayerStats_pkey" PRIMARY KEY ("id")
);

-- Unique and regular indexes for Settings
CREATE UNIQUE INDEX "Settings_guildId_key" ON "Settings"("guildId");
CREATE INDEX "Settings_guildId_channelId_idx" ON "Settings"("guildId", "channelId");

-- Indexes for Game
CREATE INDEX "Game_guildId_idx" ON "Game"("guildId");

-- Indexes for Guess
CREATE INDEX "Guess_gameId_idx" ON "Guess"("gameId");

-- Indexes for PlayerHints
CREATE INDEX "PlayerHints_userId_idx" ON "PlayerHints"("userId");

-- Indexes for PlayerStats
CREATE INDEX "PlayerStats_userId_guildId_idx" ON "PlayerStats"("userId", "guildId");

-- Foreign keys
ALTER TABLE "Game"
    ADD CONSTRAINT "Game_guildId_fkey"
    FOREIGN KEY ("guildId") REFERENCES "Settings"("guildId")
    ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "Guess"
    ADD CONSTRAINT "Guess_gameId_fkey"
    FOREIGN KEY ("gameId") REFERENCES "Game"("id")
    ON DELETE CASCADE ON UPDATE CASCADE;
