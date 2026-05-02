package publish

import (
	"context"
	"sync"
	"time"
)

// PublisherTokenRecord mirrors the Phase 1 Vault/static catalog record shape.
// The raw publish token is never represented here; only its hash is stored.
type PublisherTokenRecord struct {
	PackageName string    `json:"package" yaml:"package"`
	Subject     string    `json:"subject,omitempty" yaml:"subject,omitempty"`
	TokenHash   string    `json:"tokenHash" yaml:"token_hash"`
	CreatedAt   time.Time `json:"createdAt,omitempty" yaml:"created_at,omitempty"`
	RotatedAt   time.Time `json:"rotatedAt,omitempty" yaml:"rotated_at,omitempty"`
	RevokedAt   time.Time `json:"revokedAt,omitempty" yaml:"revoked_at,omitempty"`
	Notes       string    `json:"notes,omitempty" yaml:"notes,omitempty"`
}

// ToStaticPublisherToken converts a non-revoked catalog record into a static
// auth token entry.
func (r PublisherTokenRecord) ToStaticPublisherToken() (StaticPublisherToken, bool) {
	if !r.RevokedAt.IsZero() {
		return StaticPublisherToken{}, false
	}
	return StaticPublisherToken{Subject: r.Subject, PackageName: r.PackageName, TokenHash: r.TokenHash}, true
}

// StaticTokensFromRecords validates catalog records and returns static auth
// tokens for all non-revoked records.
func StaticTokensFromRecords(records []PublisherTokenRecord) ([]StaticPublisherToken, error) {
	tokens := make([]StaticPublisherToken, 0, len(records))
	for _, record := range records {
		token, ok := record.ToStaticPublisherToken()
		if !ok {
			continue
		}
		if err := ValidatePackageName(token.PackageName); err != nil {
			return nil, err
		}
		if _, err := NormalizeTokenHash(token.TokenHash); err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, nil
}

// PublisherCatalogSource loads package publisher records from a backing source
// such as a fixture file, Vault KV path, or later registry database.
type PublisherCatalogSource interface {
	LoadPublisherTokenRecords(ctx context.Context) ([]PublisherTokenRecord, error)
}

// ReloadablePublisherCatalog keeps the active PublisherAuth implementation in
// memory and can replace it after reloading token records from a source.
type ReloadablePublisherCatalog struct {
	mu     sync.RWMutex
	source PublisherCatalogSource
	auth   PublisherAuth
}

func NewReloadablePublisherCatalog(source PublisherCatalogSource) *ReloadablePublisherCatalog {
	return &ReloadablePublisherCatalog{source: source}
}

// Reload loads token records from the source and swaps in a new StaticTokenAuth
// only after all records validate.
func (c *ReloadablePublisherCatalog) Reload(ctx context.Context) error {
	records, err := c.source.LoadPublisherTokenRecords(ctx)
	if err != nil {
		return err
	}
	tokens, err := StaticTokensFromRecords(records)
	if err != nil {
		return err
	}
	auth, err := NewStaticTokenAuth(tokens)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.auth = auth
	return nil
}

// AuthorizePublish delegates to the currently loaded auth implementation.
func (c *ReloadablePublisherCatalog) AuthorizePublish(ctx context.Context, rawToken string, req PublishRequest) (*PublisherIdentity, error) {
	c.mu.RLock()
	auth := c.auth
	c.mu.RUnlock()
	if auth == nil {
		return nil, ErrUnauthorized
	}
	return auth.AuthorizePublish(ctx, rawToken, req)
}

// StaticPublisherCatalogSource is an in-memory source useful for tests and
// local fixtures before direct Vault loading is implemented.
type StaticPublisherCatalogSource struct {
	Records []PublisherTokenRecord
}

func (s StaticPublisherCatalogSource) LoadPublisherTokenRecords(ctx context.Context) ([]PublisherTokenRecord, error) {
	ret := make([]PublisherTokenRecord, len(s.Records))
	copy(ret, s.Records)
	return ret, nil
}
