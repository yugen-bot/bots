-- AlterTable
ALTER TABLE "Game" ADD COLUMN     "number" INTEGER NOT NULL DEFAULT 1;

-- Update
UPDATE "Game"
SET    number = g."newNumber"
FROM   (SELECT
               id AS "curId", row_number()
                 OVER (
                   PARTITION BY "guildId"
                   ORDER BY "createdAt") AS "newNumber"
        FROM   "Game") AS g
WHERE id = g."curId";
