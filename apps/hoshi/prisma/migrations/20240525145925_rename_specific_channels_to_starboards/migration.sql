-- RenameTable
ALTER TABLE "SpecificChannels" RENAME TO "Starboards";
ALTER TABLE "Starboards" RENAME CONSTRAINT "SpecificChannels_pkey" TO "Starboards_pkey";
