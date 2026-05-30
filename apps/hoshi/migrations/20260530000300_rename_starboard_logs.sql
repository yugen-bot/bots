-- Rename star_board_logs to starboard_logs to match Ent-generated table name for StarboardLog
ALTER TABLE "star_board_logs" RENAME TO "starboard_logs";

ALTER TABLE "starboard_logs" RENAME CONSTRAINT "star_board_logs_pkey" TO "starboard_logs_pkey";

ALTER INDEX "star_board_logs_message_id_key" RENAME TO "starboard_logs_message_id_key";
ALTER INDEX "star_board_logs_original_message_id_key" RENAME TO "starboard_logs_original_message_id_key";
