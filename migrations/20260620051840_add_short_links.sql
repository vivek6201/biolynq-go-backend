-- Modify "links" table
ALTER TABLE "links" DROP CONSTRAINT "fk_links_profile", ADD CONSTRAINT "fk_profiles_links" FOREIGN KEY ("profile_id") REFERENCES "profiles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Create "short_links" table
CREATE TABLE "short_links" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "link_id" uuid NOT NULL,
  "slug" character varying(100) NOT NULL,
  "is_active" boolean NOT NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_links_short_links" FOREIGN KEY ("link_id") REFERENCES "links" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_short_links_link_id" to table: "short_links"
CREATE INDEX "idx_short_links_link_id" ON "short_links" ("link_id");
-- Create index "idx_short_links_slug" to table: "short_links"
CREATE UNIQUE INDEX "idx_short_links_slug" ON "short_links" ("slug");
-- Modify "analytic_events" table
ALTER TABLE "analytic_events" ADD COLUMN "short_link_id" uuid NULL, ADD CONSTRAINT "fk_analytic_events_short_link" FOREIGN KEY ("short_link_id") REFERENCES "short_links" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Create index "idx_analytic_events_short_link_id" to table: "analytic_events"
CREATE INDEX "idx_analytic_events_short_link_id" ON "analytic_events" ("short_link_id");
