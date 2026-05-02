package publish

import (
	"context"
	"encoding/json"
	"os"
)

// FilePublisherCatalogSource loads Phase 1 publisher token records from a JSON
// file. It mirrors the Vault record shape and is useful for local smoke tests
// and operator-synced Vault exports.
type FilePublisherCatalogSource struct {
	Path string
}

type publisherCatalogFile struct {
	Publishers []PublisherTokenRecord `json:"publishers"`
}

func (s FilePublisherCatalogSource) LoadPublisherTokenRecords(ctx context.Context) ([]PublisherTokenRecord, error) {
	data, err := os.ReadFile(s.Path)
	if err != nil {
		return nil, err
	}
	var file publisherCatalogFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	ret := make([]PublisherTokenRecord, len(file.Publishers))
	copy(ret, file.Publishers)
	return ret, nil
}
