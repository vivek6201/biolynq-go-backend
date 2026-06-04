-- Create "temp_users" table
CREATE TABLE "temp_users" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "email" character varying(255) NOT NULL,
  "display_name" character varying(100) NULL,
  "avatar_url" text NULL,
  "provider" character varying(20) NULL DEFAULT 'email_otp',
  "is_expired" boolean NULL DEFAULT false,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_temp_users_email" to table: "temp_users"
CREATE UNIQUE INDEX "idx_temp_users_email" ON "temp_users" ("email");
-- Create "users" table
CREATE TABLE "users" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "email" character varying(255) NOT NULL,
  "password_hash" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_users_email" to table: "users"
CREATE UNIQUE INDEX "idx_users_email" ON "users" ("email");
-- Create "profiles" table
CREATE TABLE "profiles" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "username" character varying(32) NOT NULL,
  "display_name" character varying(100) NULL,
  "bio" text NULL,
  "avatar_url" text NULL,
  "theme" character varying(50) NULL DEFAULT 'default',
  "is_public" boolean NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_profiles_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_profiles_user_id" to table: "profiles"
CREATE UNIQUE INDEX "idx_profiles_user_id" ON "profiles" ("user_id");
-- Create index "idx_profiles_username" to table: "profiles"
CREATE UNIQUE INDEX "idx_profiles_username" ON "profiles" ("username");
-- Create "sessions" table
CREATE TABLE "sessions" (
  "id" character varying(255) NOT NULL,
  "user_id" uuid NOT NULL,
  "provider" character varying(20) NOT NULL,
  "ip_address" character varying(45) NULL,
  "user_agent" text NULL,
  "expires_at" timestamptz NOT NULL,
  "last_rotated_at" timestamptz NOT NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_sessions_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_sessions_expires_at" to table: "sessions"
CREATE INDEX "idx_sessions_expires_at" ON "sessions" ("expires_at");
-- Create index "idx_sessions_user_id" to table: "sessions"
CREATE INDEX "idx_sessions_user_id" ON "sessions" ("user_id");
