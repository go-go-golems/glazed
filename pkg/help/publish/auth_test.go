package publish

import (
	"context"
	"errors"
	"testing"
)

func TestStaticTokenAuthAllowsMatchingPackage(t *testing.T) {
	auth := newTestStaticAuth(t)
	identity, err := auth.AuthorizePublish(context.Background(), "pinocchio-secret", PublishRequest{PackageName: "pinocchio", Version: "v1"})
	if err != nil {
		t.Fatalf("AuthorizePublish: %v", err)
	}
	if identity.PackageName != "pinocchio" || identity.Method != "static-token" {
		t.Fatalf("unexpected identity: %#v", identity)
	}
}

func TestStaticTokenAuthRejectsUnknownToken(t *testing.T) {
	auth := newTestStaticAuth(t)
	_, err := auth.AuthorizePublish(context.Background(), "unknown", PublishRequest{PackageName: "pinocchio", Version: "v1"})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestStaticTokenAuthRejectsEmptyToken(t *testing.T) {
	auth := newTestStaticAuth(t)
	_, err := auth.AuthorizePublish(context.Background(), "", PublishRequest{PackageName: "pinocchio", Version: "v1"})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestStaticTokenAuthRejectsWrongPackage(t *testing.T) {
	auth := newTestStaticAuth(t)
	_, err := auth.AuthorizePublish(context.Background(), "pinocchio-secret", PublishRequest{PackageName: "glazed", Version: "v1"})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestStaticTokenAuthRejectsInvalidRequest(t *testing.T) {
	auth := newTestStaticAuth(t)
	_, err := auth.AuthorizePublish(context.Background(), "pinocchio-secret", PublishRequest{PackageName: "../pinocchio", Version: "v1"})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestNewStaticTokenAuthRejectsDuplicateTokenHashes(t *testing.T) {
	hash := HashPublishToken("shared")
	_, err := NewStaticTokenAuth([]StaticPublisherToken{
		{PackageName: "pinocchio", TokenHash: hash},
		{PackageName: "glazed", TokenHash: hash},
	})
	if err == nil {
		t.Fatalf("expected duplicate hash error")
	}
}

func TestNormalizeTokenHash(t *testing.T) {
	hash := HashPublishToken("secret")
	normalized, err := NormalizeTokenHash(hash)
	if err != nil {
		t.Fatalf("NormalizeTokenHash: %v", err)
	}
	if normalized != hash {
		t.Fatalf("expected %s, got %s", hash, normalized)
	}

	if _, err := NormalizeTokenHash("secret"); err == nil {
		t.Fatalf("expected missing prefix error")
	}
	if _, err := NormalizeTokenHash("sha256:not-hex"); err == nil {
		t.Fatalf("expected hex error")
	}
}

func TestConstantTimeTokenHashEqual(t *testing.T) {
	a := HashPublishToken("a")
	if !ConstantTimeTokenHashEqual(a, a) {
		t.Fatalf("expected equal hashes")
	}
	if ConstantTimeTokenHashEqual(a, HashPublishToken("b")) {
		t.Fatalf("expected different hashes")
	}
}

func newTestStaticAuth(t *testing.T) *StaticTokenAuth {
	t.Helper()
	auth, err := NewStaticTokenAuth([]StaticPublisherToken{
		{Subject: "repo:go-go-golems/pinocchio", PackageName: "pinocchio", TokenHash: HashPublishToken("pinocchio-secret")},
		{Subject: "repo:go-go-golems/glazed", PackageName: "glazed", TokenHash: HashPublishToken("glazed-secret")},
	})
	if err != nil {
		t.Fatalf("NewStaticTokenAuth: %v", err)
	}
	return auth
}
