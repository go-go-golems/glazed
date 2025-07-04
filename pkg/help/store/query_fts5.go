//go:build sqlite_fts5

package store

// TextSearch performs full-text search using FTS5
func TextSearch(term string) Predicate {
	return func(qc *QueryCompiler) {
		qc.AddJoin("JOIN sections_fts ON sections_fts.rowid = s.id")
		qc.AddWhere("sections_fts MATCH ?", term)
	}
}
