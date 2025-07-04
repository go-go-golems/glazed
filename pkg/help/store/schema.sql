-- SQLite schema for help system
-- This file is used for documentation and initialization

-- Main sections table
CREATE TABLE IF NOT EXISTS sections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT UNIQUE NOT NULL,
    section_type TEXT NOT NULL,
    title TEXT NOT NULL,
    sub_title TEXT,
    short TEXT,
    content TEXT,
    is_top_level BOOLEAN DEFAULT FALSE,
    is_template BOOLEAN DEFAULT FALSE,
    show_per_default BOOLEAN DEFAULT FALSE,
    order_index INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Topics table for normalization
CREATE TABLE IF NOT EXISTS topics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL
);

-- Flags table for normalization
CREATE TABLE IF NOT EXISTS flags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL
);

-- Commands table for normalization
CREATE TABLE IF NOT EXISTS commands (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL
);

-- Junction tables for many-to-many relationships
CREATE TABLE IF NOT EXISTS section_topics (
    section_id INTEGER,
    topic_id INTEGER,
    PRIMARY KEY (section_id, topic_id),
    FOREIGN KEY (section_id) REFERENCES sections(id) ON DELETE CASCADE,
    FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS section_flags (
    section_id INTEGER,
    flag_id INTEGER,
    PRIMARY KEY (section_id, flag_id),
    FOREIGN KEY (section_id) REFERENCES sections(id) ON DELETE CASCADE,
    FOREIGN KEY (flag_id) REFERENCES flags(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS section_commands (
    section_id INTEGER,
    command_id INTEGER,
    PRIMARY KEY (section_id, command_id),
    FOREIGN KEY (section_id) REFERENCES sections(id) ON DELETE CASCADE,
    FOREIGN KEY (command_id) REFERENCES commands(id) ON DELETE CASCADE
);

-- FTS5 virtual table for full-text search
CREATE VIRTUAL TABLE IF NOT EXISTS sections_fts USING fts5(
    slug,
    title,
    sub_title,
    short,
    content,
    content=sections,
    content_rowid=id
);

-- Triggers to keep FTS5 table synchronized
CREATE TRIGGER IF NOT EXISTS sections_fts_insert AFTER INSERT ON sections BEGIN
    INSERT INTO sections_fts(rowid, slug, title, sub_title, short, content) 
    VALUES (new.id, new.slug, new.title, new.sub_title, new.short, new.content);
END;

CREATE TRIGGER IF NOT EXISTS sections_fts_delete AFTER DELETE ON sections BEGIN
    INSERT INTO sections_fts(sections_fts, rowid, slug, title, sub_title, short, content) 
    VALUES('delete', old.id, old.slug, old.title, old.sub_title, old.short, old.content);
END;

CREATE TRIGGER IF NOT EXISTS sections_fts_update AFTER UPDATE ON sections BEGIN
    INSERT INTO sections_fts(sections_fts, rowid, slug, title, sub_title, short, content) 
    VALUES('delete', old.id, old.slug, old.title, old.sub_title, old.short, old.content);
    INSERT INTO sections_fts(rowid, slug, title, sub_title, short, content) 
    VALUES (new.id, new.slug, new.title, new.sub_title, new.short, new.content);
END;

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_sections_slug ON sections(slug);
CREATE INDEX IF NOT EXISTS idx_sections_type ON sections(section_type);
CREATE INDEX IF NOT EXISTS idx_sections_top_level ON sections(is_top_level);
CREATE INDEX IF NOT EXISTS idx_sections_show_default ON sections(show_per_default);
CREATE INDEX IF NOT EXISTS idx_topics_name ON topics(name);
CREATE INDEX IF NOT EXISTS idx_flags_name ON flags(name);
CREATE INDEX IF NOT EXISTS idx_commands_name ON commands(name);
