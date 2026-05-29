-- CreateTable: Settings
CREATE TABLE "Settings" (
    "id" SERIAL NOT NULL,
    "guildId" TEXT NOT NULL,
    "botUpdatesChannelId" TEXT,
    "treshold" INTEGER NOT NULL DEFAULT 3,
    "self" BOOLEAN NOT NULL DEFAULT false,
    "ignoredChannelIds" TEXT[] DEFAULT ARRAY[]::TEXT[],
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "Settings_pkey" PRIMARY KEY ("id")
);

-- CreateTable: Starboards
CREATE TABLE "Starboards" (
    "id" SERIAL NOT NULL,
    "guildId" TEXT NOT NULL,
    "sourceEmoji" TEXT NOT NULL DEFAULT '⭐',
    "sourceChannelId" TEXT,
    "targetChannelId" TEXT NOT NULL,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "Starboards_pkey" PRIMARY KEY ("id")
);

-- CreateTable: Log
CREATE TABLE "Log" (
    "id" SERIAL NOT NULL,
    "guildId" TEXT NOT NULL,
    "channelId" TEXT NOT NULL,
    "messageId" TEXT NOT NULL,
    "originalMessageId" TEXT NOT NULL,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "Log_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE UNIQUE INDEX "Settings_guildId_key" ON "Settings"("guildId");

-- CreateIndex
CREATE UNIQUE INDEX "Log_messageId_key" ON "Log"("messageId");

-- CreateIndex
CREATE UNIQUE INDEX "Log_originalMessageId_key" ON "Log"("originalMessageId");
