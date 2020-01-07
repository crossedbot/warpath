
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS "captured_frames" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "type" VARCHAR(255),
    "latitude" REAL,
    "longitude" REAL,
    "distance" REAL,
    "timestamp" INTEGER,
    "source" INTEGER NOT NULL,
    "data" BLOB
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE "captured_frames";
