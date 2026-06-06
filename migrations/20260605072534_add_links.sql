-- Create "links" table
CREATE TABLE "links" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "profile_id" uuid NOT NULL,
  "title" character varying(255) NOT NULL,
  "description" text NULL,
  "url" character varying(255) NOT NULL,
  "icon_url" text NOT NULL,
  "position" bigint NOT NULL DEFAULT 0,
  "is_active" boolean NOT NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_links_profile" FOREIGN KEY ("profile_id") REFERENCES "profiles" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "link_click_events" table
CREATE TABLE "link_click_events" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "link_id" uuid NOT NULL,
  "country" character varying(100) NULL,
  "city" character varying(100) NULL,
  "browser" character varying(100) NULL,
  "os" character varying(100) NULL,
  "device" character varying(100) NULL,
  "referrer" text NULL,
  "clicked_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_link_click_events_link" FOREIGN KEY ("link_id") REFERENCES "links" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_link_click_events_clicked_at" to table: "link_click_events"
CREATE INDEX "idx_link_click_events_clicked_at" ON "link_click_events" ("clicked_at");
-- Create index "idx_link_click_events_link_id" to table: "link_click_events"
CREATE INDEX "idx_link_click_events_link_id" ON "link_click_events" ("link_id");
-- Create "link_stats" table
CREATE TABLE "link_stats" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "link_id" uuid NOT NULL,
  "total_clicks" bigint NULL DEFAULT 0,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_link_stats_link" FOREIGN KEY ("link_id") REFERENCES "links" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_link_stats_link_id" to table: "link_stats"
CREATE UNIQUE INDEX "idx_link_stats_link_id" ON "link_stats" ("link_id");
-- Create "social_links" table
CREATE TABLE "social_links" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "profile_id" uuid NOT NULL,
  "platform" character varying(50) NULL,
  "url" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_social_links_profile" FOREIGN KEY ("profile_id") REFERENCES "profiles" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_social_links_profile_id" to table: "social_links"
CREATE INDEX "idx_social_links_profile_id" ON "social_links" ("profile_id");
