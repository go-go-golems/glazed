package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	_ "github.com/mattn/go-sqlite3"
)

// Store represents a SQLite-backed help section store
type Store struct {
	db *sql.DB
}

// NewStore creates a new Store instance
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open database")
	}

	store := &Store{db: db}
	
	if err := store.initializeSchema(); err != nil {
		return nil, errors.Wrap(err, "failed to initialize schema")
	}

	return store, nil
}

// NewInMemoryStore creates a new in-memory SQLite store
func NewInMemoryStore() (*Store, error) {
	return NewStore(":memory:")
}

// initializeSchema creates the database schema
func (s *Store) initializeSchema() error {
	// Execute schema SQL statements
	for _, stmt := range schemaStatements {
		_, err := s.db.Exec(stmt)
		if err != nil {
			return errors.Wrapf(err, "failed to execute schema statement: %s", stmt)
		}
	}
	
	return nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// Find executes a query using the predicate DSL
func (s *Store) Find(ctx context.Context, predicate Predicate) ([]*help.Section, error) {
	compiler := newCompiler()
	predicate(compiler)
	
	query, args := compiler.compile()
	
	log.Debug().
		Str("query", query).
		Interface("args", args).
		Msg("Executing query")
	
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}
	defer rows.Close()
	
	var sections []*Section
	for rows.Next() {
		section := &Section{}
		var sectionTypeStr string
		err := rows.Scan(
			&section.ID,
			&section.Slug,
			&sectionTypeStr,
			&section.Title,
			&section.SubTitle,
			&section.Short,
			&section.Content,
			&section.IsTopLevel,
			&section.IsTemplate,
			&section.ShowPerDefault,
			&section.OrderIndex,
			&section.CreatedAt,
			&section.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}
		
		// Convert string to SectionType
		sectionType, err := help.SectionTypeFromString(sectionTypeStr)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid section type: %s", sectionTypeStr)
		}
		section.SectionType = sectionType
		
		sections = append(sections, section)
	}
	
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating rows")
	}
	
	// Load associated topics, flags, and commands for each section
	for _, section := range sections {
		if err := s.loadAssociatedData(ctx, section); err != nil {
			return nil, errors.Wrapf(err, "failed to load associated data for section %s", section.Slug)
		}
	}
	
	// Convert to help.Section
	result := make([]*help.Section, len(sections))
	for i, section := range sections {
		result[i] = section.ToHelpSection()
	}
	
	return result, nil
}

// loadAssociatedData loads topics, flags, and commands for a section
func (s *Store) loadAssociatedData(ctx context.Context, section *Section) error {
	// Load topics
	topicsQuery := `
		SELECT t.name 
		FROM topics t 
		JOIN section_topics st ON t.id = st.topic_id 
		WHERE st.section_id = ?
	`
	topics, err := s.loadStringSlice(ctx, topicsQuery, section.ID)
	if err != nil {
		return errors.Wrap(err, "failed to load topics")
	}
	section.Topics = topics
	
	// Load flags
	flagsQuery := `
		SELECT f.name 
		FROM flags f 
		JOIN section_flags sf ON f.id = sf.flag_id 
		WHERE sf.section_id = ?
	`
	flags, err := s.loadStringSlice(ctx, flagsQuery, section.ID)
	if err != nil {
		return errors.Wrap(err, "failed to load flags")
	}
	section.Flags = flags
	
	// Load commands
	commandsQuery := `
		SELECT c.name 
		FROM commands c 
		JOIN section_commands sc ON c.id = sc.command_id 
		WHERE sc.section_id = ?
	`
	commands, err := s.loadStringSlice(ctx, commandsQuery, section.ID)
	if err != nil {
		return errors.Wrap(err, "failed to load commands")
	}
	section.Commands = commands
	
	return nil
}

// loadStringSlice is a helper function to load a slice of strings
func (s *Store) loadStringSlice(ctx context.Context, query string, id int64) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var result []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	
	return result, rows.Err()
}

// AddSection adds or updates a section in the store
func (s *Store) AddSection(ctx context.Context, section *help.Section) error {
	return s.addSection(ctx, FromHelpSection(section))
}

// addSection adds or updates a section in the store
func (s *Store) addSection(ctx context.Context, section *Section) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()
	
	// Insert or update section
	sectionQuery := `
		INSERT OR REPLACE INTO sections 
		(slug, section_type, title, sub_title, short, content, is_top_level, is_template, show_per_default, order_index) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := tx.ExecContext(ctx, sectionQuery, 
		section.Slug,
		section.SectionType.String(),
		section.Title,
		section.SubTitle,
		section.Short,
		section.Content,
		section.IsTopLevel,
		section.IsTemplate,
		section.ShowPerDefault,
		section.OrderIndex,
	)
	if err != nil {
		return errors.Wrap(err, "failed to insert section")
	}
	
	// Get the section ID
	var sectionID int64
	if section.ID == 0 {
		sectionID, err = result.LastInsertId()
		if err != nil {
			return errors.Wrap(err, "failed to get section ID")
		}
	} else {
		sectionID = section.ID
	}
	
	// Clear existing associations
	if err := s.clearAssociations(ctx, tx, sectionID); err != nil {
		return errors.Wrap(err, "failed to clear associations")
	}
	
	// Insert topics
	if err := s.insertTopics(ctx, tx, sectionID, section.Topics); err != nil {
		return errors.Wrap(err, "failed to insert topics")
	}
	
	// Insert flags
	if err := s.insertFlags(ctx, tx, sectionID, section.Flags); err != nil {
		return errors.Wrap(err, "failed to insert flags")
	}
	
	// Insert commands
	if err := s.insertCommands(ctx, tx, sectionID, section.Commands); err != nil {
		return errors.Wrap(err, "failed to insert commands")
	}
	
	return tx.Commit()
}

// clearAssociations removes all topic, flag, and command associations for a section
func (s *Store) clearAssociations(ctx context.Context, tx *sql.Tx, sectionID int64) error {
	queries := []string{
		"DELETE FROM section_topics WHERE section_id = ?",
		"DELETE FROM section_flags WHERE section_id = ?",
		"DELETE FROM section_commands WHERE section_id = ?",
	}
	
	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query, sectionID); err != nil {
			return err
		}
	}
	
	return nil
}

// insertTopics inserts topics and their associations
func (s *Store) insertTopics(ctx context.Context, tx *sql.Tx, sectionID int64, topics []string) error {
	for _, topic := range topics {
		// Insert topic if not exists
		_, err := tx.ExecContext(ctx, "INSERT OR IGNORE INTO topics (name) VALUES (?)", topic)
		if err != nil {
			return err
		}
		
		// Get topic ID
		var topicID int64
		err = tx.QueryRowContext(ctx, "SELECT id FROM topics WHERE name = ?", topic).Scan(&topicID)
		if err != nil {
			return err
		}
		
		// Insert association
		_, err = tx.ExecContext(ctx, 
			"INSERT OR IGNORE INTO section_topics (section_id, topic_id) VALUES (?, ?)", 
			sectionID, topicID)
		if err != nil {
			return err
		}
	}
	
	return nil
}

// insertFlags inserts flags and their associations
func (s *Store) insertFlags(ctx context.Context, tx *sql.Tx, sectionID int64, flags []string) error {
	for _, flag := range flags {
		// Insert flag if not exists
		_, err := tx.ExecContext(ctx, "INSERT OR IGNORE INTO flags (name) VALUES (?)", flag)
		if err != nil {
			return err
		}
		
		// Get flag ID
		var flagID int64
		err = tx.QueryRowContext(ctx, "SELECT id FROM flags WHERE name = ?", flag).Scan(&flagID)
		if err != nil {
			return err
		}
		
		// Insert association
		_, err = tx.ExecContext(ctx, 
			"INSERT OR IGNORE INTO section_flags (section_id, flag_id) VALUES (?, ?)", 
			sectionID, flagID)
		if err != nil {
			return err
		}
	}
	
	return nil
}

// insertCommands inserts commands and their associations
func (s *Store) insertCommands(ctx context.Context, tx *sql.Tx, sectionID int64, commands []string) error {
	for _, command := range commands {
		// Insert command if not exists
		_, err := tx.ExecContext(ctx, "INSERT OR IGNORE INTO commands (name) VALUES (?)", command)
		if err != nil {
			return err
		}
		
		// Get command ID
		var commandID int64
		err = tx.QueryRowContext(ctx, "SELECT id FROM commands WHERE name = ?", command).Scan(&commandID)
		if err != nil {
			return err
		}
		
		// Insert association
		_, err = tx.ExecContext(ctx, 
			"INSERT OR IGNORE INTO section_commands (section_id, command_id) VALUES (?, ?)", 
			sectionID, commandID)
		if err != nil {
			return err
		}
	}
	
	return nil
}

// GetSectionBySlug retrieves a section by its slug
func (s *Store) GetSectionBySlug(ctx context.Context, slug string) (*help.Section, error) {
	sections, err := s.Find(ctx, SlugEquals(slug))
	if err != nil {
		return nil, err
	}
	
	if len(sections) == 0 {
		return nil, fmt.Errorf("section with slug %s not found", slug)
	}
	
	return sections[0], nil
}

// LoadSectionsFromHelpSystem loads all sections from a help system
func (s *Store) LoadSectionsFromHelpSystem(ctx context.Context, hs *help.HelpSystem) error {
	for _, section := range hs.Sections {
		if err := s.AddSection(ctx, section); err != nil {
			return errors.Wrapf(err, "failed to add section %s", section.Slug)
		}
	}
	
	return nil
}

// GetAllSections retrieves all sections
func (s *Store) GetAllSections(ctx context.Context) ([]*help.Section, error) {
	return s.Find(ctx, func(c *compiler) {
		// No additional filters - return all sections
	})
}

// GetStats returns statistics about the store
func (s *Store) GetStats(ctx context.Context) (map[string]int, error) {
	stats := make(map[string]int)
	
	queries := map[string]string{
		"sections": "SELECT COUNT(*) FROM sections",
		"topics":   "SELECT COUNT(*) FROM topics",
		"flags":    "SELECT COUNT(*) FROM flags",
		"commands": "SELECT COUNT(*) FROM commands",
	}
	
	for key, query := range queries {
		var count int
		err := s.db.QueryRowContext(ctx, query).Scan(&count)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get %s count", key)
		}
		stats[key] = count
	}
	
	return stats, nil
}
