//go:build !sqlite_fts5

package store

// createFTSTables is a no-op when FTS5 is not enabled
func (s *Store) createFTSTables() error {
	// log.Trace().Msg("FTS5 support disabled, skipping FTS table creation")
	return nil
}
