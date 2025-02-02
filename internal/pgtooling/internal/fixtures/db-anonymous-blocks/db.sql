
CREATE TABLE branches (
  id TEXT PRIMARY KEY,
  "default" boolean,
  "primary" boolean
);

INSERT INTO branches (id, "default", "primary") VALUES
  ('one', false, true),
  ('two', false, true),
  ('three', false, true),
  ('four', true, true);
