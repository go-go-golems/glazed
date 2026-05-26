package publish

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

// JWTPublisherAuth authorizes package publishes with OIDC-compliant JWTs.
//
// This auth mode is intended for Vault Identity/OIDC publish tokens. Vault
// signs a short-lived token with audience "docs-registry" and a package claim.
// The registry validates the token through OIDC discovery/JWKS and then checks
// that the package claim matches the requested package.
type JWTPublisherAuth struct {
	Issuer   string
	ClientID string

	verifier *oidc.IDTokenVerifier
}

// NewJWTPublisherAuth constructs a JWT publisher auth implementation using
// OIDC discovery from issuer. clientID is the expected JWT audience.
func NewJWTPublisherAuth(ctx context.Context, issuer, clientID string) (*JWTPublisherAuth, error) {
	issuer = strings.TrimSpace(issuer)
	clientID = strings.TrimSpace(clientID)
	if issuer == "" {
		return nil, errors.New("jwt issuer must not be empty")
	}
	if clientID == "" {
		return nil, errors.New("jwt client id must not be empty")
	}

	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("load OIDC provider metadata: %w", err)
	}

	return &JWTPublisherAuth{
		Issuer:   issuer,
		ClientID: clientID,
		verifier: provider.Verifier(&oidc.Config{ClientID: clientID}),
	}, nil
}

type docsPublishJWTClaims struct {
	TokenUse       string `json:"token_use"`
	PackageName    string `json:"package"`
	Repository     string `json:"repository"`
	RepositoryID   string `json:"repository_id"`
	WorkflowRef    string `json:"workflow_ref"`
	JobWorkflowRef string `json:"job_workflow_ref"`
	RunID          string `json:"run_id"`
}

// AuthorizePublish validates rawToken as an OIDC JWT and authorizes it for req.
func (a *JWTPublisherAuth) AuthorizePublish(ctx context.Context, rawToken string, req PublishRequest) (*PublisherIdentity, error) {
	if strings.TrimSpace(rawToken) == "" {
		return nil, ErrUnauthorized
	}
	if err := ValidatePackageVersion(req.PackageName, req.Version); err != nil {
		return nil, err
	}
	if a == nil || a.verifier == nil {
		return nil, errors.New("jwt publisher auth is not configured")
	}

	idToken, err := a.verifier.Verify(ctx, rawToken)
	if err != nil {
		return nil, ErrUnauthorized
	}

	var claims docsPublishJWTClaims
	if err := idToken.Claims(&claims); err != nil {
		return nil, ErrUnauthorized
	}

	if claims.TokenUse != "docsctl-publish" {
		return nil, ErrForbidden
	}
	if claims.PackageName == "" || claims.PackageName != req.PackageName {
		return nil, ErrForbidden
	}

	subject := claims.Repository
	if subject == "" {
		subject = idToken.Subject
	}
	if subject == "" {
		subject = claims.PackageName
	}

	return &PublisherIdentity{Subject: subject, PackageName: claims.PackageName, Method: "vault-oidc-jwt"}, nil
}
