package publish

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestStaticTokensFromRecordsSkipsRevoked(t *testing.T) {
	records := []PublisherTokenRecord{
		{PackageName: "pinocchio", TokenHash: HashPublishToken("pinocchio")},
		{PackageName: "glazed", TokenHash: HashPublishToken("glazed"), RevokedAt: time.Now()},
	}
	tokens, err := StaticTokensFromRecords(records)
	if err != nil {
		t.Fatalf("StaticTokensFromRecords: %v", err)
	}
	if len(tokens) != 1 || tokens[0].PackageName != "pinocchio" {
		t.Fatalf("unexpected tokens: %#v", tokens)
	}
}

func TestStaticTokensFromRecordsValidatesRecords(t *testing.T) {
	_, err := StaticTokensFromRecords([]PublisherTokenRecord{{PackageName: "../bad", TokenHash: HashPublishToken("bad")}})
	if err == nil {
		t.Fatalf("expected invalid package error")
	}

	_, err = StaticTokensFromRecords([]PublisherTokenRecord{{PackageName: "pinocchio", TokenHash: "bad"}})
	if err == nil {
		t.Fatalf("expected invalid hash error")
	}
}

func TestReloadablePublisherCatalog(t *testing.T) {
	source := StaticPublisherCatalogSource{Records: []PublisherTokenRecord{{PackageName: "pinocchio", TokenHash: HashPublishToken("pinocchio-token")}}}
	catalog := NewReloadablePublisherCatalog(source)

	_, err := catalog.AuthorizePublish(context.Background(), "pinocchio-token", PublishRequest{PackageName: "pinocchio", Version: "v1"})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected unauthorized before reload, got %v", err)
	}

	if err := catalog.Reload(context.Background()); err != nil {
		t.Fatalf("Reload: %v", err)
	}
	identity, err := catalog.AuthorizePublish(context.Background(), "pinocchio-token", PublishRequest{PackageName: "pinocchio", Version: "v1"})
	if err != nil {
		t.Fatalf("AuthorizePublish: %v", err)
	}
	if identity.PackageName != "pinocchio" {
		t.Fatalf("unexpected identity: %#v", identity)
	}
}

type failingCatalogSource struct{}

func (f failingCatalogSource) LoadPublisherTokenRecords(ctx context.Context) ([]PublisherTokenRecord, error) {
	return nil, errors.New("boom")
}

func TestReloadablePublisherCatalogKeepsPreviousAuthOnReloadFailure(t *testing.T) {
	catalog := NewReloadablePublisherCatalog(StaticPublisherCatalogSource{Records: []PublisherTokenRecord{{PackageName: "pinocchio", TokenHash: HashPublishToken("pinocchio-token")}}})
	if err := catalog.Reload(context.Background()); err != nil {
		t.Fatalf("initial reload: %v", err)
	}

	catalog.source = failingCatalogSource{}
	if err := catalog.Reload(context.Background()); err == nil {
		t.Fatalf("expected reload failure")
	}
	_, err := catalog.AuthorizePublish(context.Background(), "pinocchio-token", PublishRequest{PackageName: "pinocchio", Version: "v1"})
	if err != nil {
		t.Fatalf("previous auth should still work after failed reload: %v", err)
	}
}
