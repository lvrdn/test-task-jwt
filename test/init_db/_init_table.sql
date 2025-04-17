DROP TABLE IF EXISTS "auth";
CREATE TABLE auth (
    "id" serial PRIMARY KEY,
    "guid" uuid NOT NULL,
    "hashed_refresh_token"  bytea,
    "exp_date" date
);

CREATE INDEX idx_auth_guid_hash ON auth USING HASH (guid);

INSERT INTO "auth" ("id", "guid") VALUES
(1,'2b421a0e-c5fa-47e3-a9d9-bc2fcf31ffe6');