-- Modify "links" table
ALTER TABLE "links" ADD COLUMN "is_social" boolean NOT NULL DEFAULT false;
-- Create "visitor_metadata" table
CREATE TABLE "visitor_metadata" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "ip" character varying(45) NOT NULL,
  "country" character varying(100) NULL,
  "city" character varying(100) NULL,
  "browser" character varying(100) NOT NULL,
  "os" character varying(100) NOT NULL,
  "device" character varying(100) NOT NULL,
  "user_agent" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_visitor_client" to table: "visitor_metadata"
CREATE UNIQUE INDEX "idx_visitor_client" ON "visitor_metadata" ("ip", "browser", "os", "device");
-- Create "analytic_events" table
CREATE TABLE "analytic_events" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "profile_id" uuid NOT NULL,
  "event_type" character varying(50) NOT NULL,
  "link_id" uuid NULL,
  "visitor_metadata_id" uuid NOT NULL,
  "referrer" text NULL,
  "clicked_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_analytic_events_link" FOREIGN KEY ("link_id") REFERENCES "links" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "fk_analytic_events_profile" FOREIGN KEY ("profile_id") REFERENCES "profiles" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "fk_analytic_events_visitor_metadata" FOREIGN KEY ("visitor_metadata_id") REFERENCES "visitor_metadata" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_analytic_events_clicked_at" to table: "analytic_events"
CREATE INDEX "idx_analytic_events_clicked_at" ON "analytic_events" ("clicked_at");
-- Create index "idx_analytic_events_event_type" to table: "analytic_events"
CREATE INDEX "idx_analytic_events_event_type" ON "analytic_events" ("event_type");
-- Create index "idx_analytic_events_link_id" to table: "analytic_events"
CREATE INDEX "idx_analytic_events_link_id" ON "analytic_events" ("link_id");
-- Create index "idx_analytic_events_profile_id" to table: "analytic_events"
CREATE INDEX "idx_analytic_events_profile_id" ON "analytic_events" ("profile_id");
-- Create index "idx_analytic_events_visitor_metadata_id" to table: "analytic_events"
CREATE INDEX "idx_analytic_events_visitor_metadata_id" ON "analytic_events" ("visitor_metadata_id");
-- Drop "link_click_events" table
DROP TABLE "link_click_events";
-- Drop "link_stats" table
DROP TABLE "link_stats";
-- Drop "social_links" table
DROP TABLE "social_links";
