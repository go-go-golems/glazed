package publish

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const tokenHashPrefix = "sha256:"

var (
	// ErrUnauthorized means the caller did not present a recognized publish token.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden means the token is known but not allowed for the requested package/version.
	ErrForbidden = errors.New("forbidden")
)

// PublishRequest is the authorization request for publishing one package version.
type PublishRequest struct {
	PackageName string
	Version     string
}

// PublisherIdentity describes the authenticated publisher after authorization.
type PublisherIdentity struct {
	Subject     string
	PackageName string
	Method      string
}

// PublisherAuth authorizes package/version publishing requests.
type PublisherAuth interface {
	AuthorizePublish(ctx context.Context, rawToken string, req PublishRequest) (*PublisherIdentity, error)
}

// StaticPublisherToken binds one token hash to exactly one package.
type StaticPublisherToken struct {
	Subject     string
	PackageName string
	TokenHash   string
}

// StaticTokenAuth authorizes publishes against an in-memory package token list.
type StaticTokenAuth struct {
	tokens []StaticPublisherToken
}

// NewStaticTokenAuth validates and constructs static package-token auth.
func NewStaticTokenAuth(tokens []StaticPublisherToken) (*StaticTokenAuth, error) {
	seenHashes := map[string]struct{}{}
	ret := make([]StaticPublisherToken, 0, len(tokens))
	for _, token := range tokens {
		if err := ValidatePackageName(token.PackageName); err != nil {
			return nil, err
		}
		normalizedHash, err := NormalizeTokenHash(token.TokenHash)
		if err != nil {
			return nil, err
		}
		if _, ok := seenHashes[normalizedHash]; ok {
			return nil, fmt.Errorf("duplicate publisher token hash for package %q", token.PackageName)
		}
		seenHashes[normalizedHash] = struct{}{}
		token.TokenHash = normalizedHash
		if token.Subject == "" {
			token.Subject = token.PackageName
		}
		ret = append(ret, token)
	}
	return &StaticTokenAuth{tokens: ret}, nil
}

// AuthorizePublish authorizes rawToken for req. It intentionally compares the
// presented token hash with stored hashes using constant-time comparison.
func (a *StaticTokenAuth) AuthorizePublish(ctx context.Context, rawToken string, req PublishRequest) (*PublisherIdentity, error) {
	if rawToken == "" {
		return nil, ErrUnauthorized
	}
	if err := ValidatePackageVersion(req.PackageName, req.Version); err != nil {
		return nil, err
	}

	presentedHash := HashPublishToken(rawToken)
	var matched *StaticPublisherToken
	for i := range a.tokens {
		if ConstantTimeTokenHashEqual(a.tokens[i].TokenHash, presentedHash) {
			matched = &a.tokens[i]
		}
	}
	if matched == nil {
		return nil, ErrUnauthorized
	}
	if matched.PackageName != req.PackageName {
		return nil, ErrForbidden
	}
	return &PublisherIdentity{Subject: matched.Subject, PackageName: matched.PackageName, Method: "static-token"}, nil
}

// HashPublishToken returns the canonical hash form stored for a raw publish token.
func HashPublishToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return tokenHashPrefix + hex.EncodeToString(sum[:])
}

// NormalizeTokenHash validates and normalizes a stored token hash.
func NormalizeTokenHash(tokenHash string) (string, error) {
	tokenHash = strings.TrimSpace(tokenHash)
	if tokenHash == "" {
		return "", errors.New("token hash must not be empty")
	}
	if !strings.HasPrefix(tokenHash, tokenHashPrefix) {
		return "", fmt.Errorf("token hash must use %s prefix", tokenHashPrefix)
	}
	hexPart := strings.TrimPrefix(tokenHash, tokenHashPrefix)
	decoded, err := hex.DecodeString(hexPart)
	if err != nil {
		return "", fmt.Errorf("token hash must contain hex sha256 digest: %w", err)
	}
	if len(decoded) != sha256.Size {
		return "", fmt.Errorf("token hash digest must be %d bytes", sha256.Size)
	}
	return tokenHashPrefix + strings.ToLower(hexPart), nil
}

// ConstantTimeTokenHashEqual compares canonical token hash strings without
// leaking which byte differed.
func ConstantTimeTokenHashEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
