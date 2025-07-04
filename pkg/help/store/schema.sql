CREATE TABLE IF NOT EXISTS sections (
  id          INTEGER PRIMARY KEY,
  slug        TEXT UNIQUE NOT NULL,
  title       TEXT,
  subtitle    TEXT,
  short       TEXT,
  content     TEXT,
  sectionType TEXT,
  isTopLevel  BOOLEAN,
  isTemplate  BOOLEAN,
  showDefault BOOLEAN,
  ord         INTEGER
);
CREATE TABLE IF NOT EXISTS section_topics   (section_id INTEGER, topic   TEXT);
CREATE TABLE IF NOT EXISTS section_flags    (section_id INTEGER, flag    TEXT);
CREATE TABLE IF NOT EXISTS section_commands (section_id INTEGER, command TEXT);
CREATE VIRTUAL TABLE IF NOT EXISTS section_fts USING fts5(
  slug, title, subtitle, short, content, content='sections'
);
CREATE INDEX IF NOT EXISTS idx_topics_topic       ON section_topics(topic);
CREATE INDEX IF NOT EXISTS idx_flags_flag         ON section_flags(flag);
CREATE INDEX IF NOT EXISTS idx_commands_command   ON section_commands(command);
CREATE INDEX IF NOT EXISTS idx_sections_type      ON sections(sectionType);
CREATE INDEX IF NOT EXISTS idx_sections_toplevel  ON sections(isTopLevel); 