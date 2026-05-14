-- CreateTable
CREATE TABLE "Settings" (
    "id" SERIAL NOT NULL,
    "guildId" TEXT NOT NULL,

    CONSTRAINT "Settings_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE UNIQUE INDEX "Settings_guildId_key" ON "Settings"("guildId");
