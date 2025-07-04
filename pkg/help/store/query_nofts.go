//go:build !sqlite_fts5

package store

// TextSearch performs a fallback text search using LIKE when FTS5 is not available
func TextSearch(term string) Predicate {
	return func(qc *QueryCompiler) {
		// Fallback to LIKE search in title and content
		qc.AddWhere("(s.title LIKE ? OR s.content LIKE ?)", "%"+term+"%", "%"+term+"%")
	}
}
