package store

const schema = `
CREATE TABLE IF NOT EXISTS sections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT UNIQUE NOT NULL,
    title TEXT,
    subtitle TEXT,
    short TEXT,
    content TEXT,
    sectionType TEXT,
    isTopLevel BOOLEAN DEFAULT FALSE,
    isTemplate BOOLEAN DEFAULT FALSE,
    showDefault BOOLEAN DEFAULT FALSE,
    ord INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS section_topics (
    section_id INTEGER,
    topic TEXT,
    FOREIGN KEY (section_id) REFERENCES sections(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS section_flags (
    section_id INTEGER,
    flag TEXT,
    FOREIGN KEY (section_id) REFERENCES sections(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS section_commands (
    section_id INTEGER,
    command TEXT,
    FOREIGN KEY (section_id) REFERENCES sections(id) ON DELETE CASCADE
);

CREATE VIRTUAL TABLE IF NOT EXISTS section_fts USING fts5(
    slug, title, subtitle, short, content, content='sections', content_rowid='id'
);

CREATE INDEX IF NOT EXISTS idx_sections_slug ON sections(slug);
CREATE INDEX IF NOT EXISTS idx_sections_type ON sections(sectionType);
CREATE INDEX IF NOT EXISTS idx_sections_toplevel ON sections(isTopLevel);
CREATE INDEX IF NOT EXISTS idx_sections_showdefault ON sections(showDefault);
CREATE INDEX IF NOT EXISTS idx_sections_ord ON sections(ord);

CREATE INDEX IF NOT EXISTS idx_topics_section_id ON section_topics(section_id);
CREATE INDEX IF NOT EXISTS idx_topics_topic ON section_topics(topic);
CREATE INDEX IF NOT EXISTS idx_flags_section_id ON section_flags(section_id);
CREATE INDEX IF NOT EXISTS idx_flags_flag ON section_flags(flag);
CREATE INDEX IF NOT EXISTS idx_commands_section_id ON section_commands(section_id);
CREATE INDEX IF NOT EXISTS idx_commands_command ON section_commands(command);

-- Triggers to maintain FTS table
CREATE TRIGGER IF NOT EXISTS section_fts_insert AFTER INSERT ON sections BEGIN
    INSERT INTO section_fts(rowid, slug, title, subtitle, short, content) 
    VALUES (new.id, new.slug, new.title, new.subtitle, new.short, new.content);
END;

CREATE TRIGGER IF NOT EXISTS section_fts_delete AFTER DELETE ON sections BEGIN
    INSERT INTO section_fts(section_fts, rowid, slug, title, subtitle, short, content) 
    VALUES('delete', old.id, old.slug, old.title, old.subtitle, old.short, old.content);
END;

CREATE TRIGGER IF NOT EXISTS section_fts_update AFTER UPDATE ON sections BEGIN
    INSERT INTO section_fts(section_fts, rowid, slug, title, subtitle, short, content) 
    VALUES('delete', old.id, old.slug, old.title, old.subtitle, old.short, old.content);
    INSERT INTO section_fts(rowid, slug, title, subtitle, short, content) 
    VALUES (new.id, new.slug, new.title, new.subtitle, new.short, new.content);
END;
`
