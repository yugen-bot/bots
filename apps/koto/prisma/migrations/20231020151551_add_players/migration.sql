-- CreateTable
CREATE TABLE "Players" (
    "id" SERIAL NOT NULL,
    "userId" TEXT NOT NULL,
    "guildId" TEXT NOT NULL,
    "points" INTEGER NOT NULL DEFAULT 0,
    "participated" INTEGER NOT NULL DEFAULT 0,
    "wins" INTEGER NOT NULL DEFAULT 0,

    CONSTRAINT "Players_pkey" PRIMARY KEY ("id")
);
