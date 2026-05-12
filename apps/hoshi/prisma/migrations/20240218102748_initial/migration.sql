-- CreateTable
CREATE TABLE "Settings" (
    "id" SERIAL NOT NULL,
    "guildId" TEXT NOT NULL,
    "enabled" BOOLEAN NOT NULL DEFAULT true,
    "channelId" TEXT,
    "treshold" INTEGER NOT NULL DEFAULT 3,
    "emoji" TEXT NOT NULL DEFAULT '⭐',
    "self" BOOLEAN NOT NULL DEFAULT false,
    "ignoredChannelIds" TEXT[] DEFAULT ARRAY[]::TEXT[],
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "Settings_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "SpecificChannels" (
    "id" SERIAL NOT NULL,
    "guildId" TEXT NOT NULL,
    "sourceChannelId" TEXT NOT NULL,
    "channelId" TEXT NOT NULL,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "SpecificChannels_pkey" PRIMARY KEY ("id")
);

-- CreateTable
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
