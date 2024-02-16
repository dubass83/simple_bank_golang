CREATE TABLE "sessions" (
  "id" uuid PRIMARY KEY NOT NULL,
  "username" varchar NOT NULL,
  "refresh_token" varchar NOT NULL,
  "user_agent" varchar NOT NULL,
  "client_ip" varchar NOT NULL,
  "is_bloked" boolean NOT NULL DEFAULT false, 
  "expired_at" timestamptz NOT NULL ,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

ALTER TABLE "sessions" ADD FOREIGN KEY ("username") REFERENCES "users" ("username");