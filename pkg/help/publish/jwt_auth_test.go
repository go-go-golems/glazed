package publish

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type testOIDCIssuer struct {
	issuer string
	key    *rsa.PrivateKey
	kid    string
	server *httptest.Server
}

func newTestOIDCIssuer(t *testing.T) *testOIDCIssuer {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	ret := &testOIDCIssuer{key: key, kid: "test-key"}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	ret.server = server
	ret.issuer = server.URL

	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		writeTestJSON(t, w, map[string]any{
			"issuer":                                ret.issuer,
			"jwks_uri":                              ret.issuer + "/keys",
			"response_types_supported":              []string{"id_token"},
			"subject_types_supported":               []string{"public"},
			"id_token_signing_alg_values_supported": []string{"RS256"},
		})
	})
	mux.HandleFunc("/keys", func(w http.ResponseWriter, r *http.Request) {
		writeTestJSON(t, w, map[string]any{
			"keys": []map[string]any{rsaPublicJWK(&key.PublicKey, ret.kid)},
		})
	})

	t.Cleanup(server.Close)
	return ret
}

func writeTestJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatalf("write json: %v", err)
	}
}

func rsaPublicJWK(pub *rsa.PublicKey, kid string) map[string]any {
	return map[string]any{
		"kty": "RSA",
		"use": "sig",
		"kid": kid,
		"alg": "RS256",
		"n":   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
	}
}

func (i *testOIDCIssuer) sign(t *testing.T, claims map[string]any) string {
	t.Helper()
	now := time.Now()
	if _, ok := claims["iss"]; !ok {
		claims["iss"] = i.issuer
	}
	if _, ok := claims["sub"]; !ok {
		claims["sub"] = "vault-entity-123"
	}
	if _, ok := claims["aud"]; !ok {
		claims["aud"] = "docs-registry"
	}
	if _, ok := claims["iat"]; !ok {
		claims["iat"] = now.Unix()
	}
	if _, ok := claims["exp"]; !ok {
		claims["exp"] = now.Add(5 * time.Minute).Unix()
	}

	header := map[string]any{"alg": "RS256", "kid": i.kid, "typ": "JWT"}
	encodedHeader := mustBase64JSON(t, header)
	encodedClaims := mustBase64JSON(t, claims)
	signingInput := encodedHeader + "." + encodedClaims
	digest := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, i.key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig)
}

func mustBase64JSON(t *testing.T, v any) string {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return base64.RawURLEncoding.EncodeToString(data)
}

func newTestJWTAuth(t *testing.T, issuer *testOIDCIssuer) *JWTPublisherAuth {
	t.Helper()
	auth, err := NewJWTPublisherAuth(context.Background(), issuer.issuer, "docs-registry")
	if err != nil {
		t.Fatalf("new jwt auth: %v", err)
	}
	return auth
}

func validDocsPublishClaims(packageName string) map[string]any {
	return map[string]any{
		"token_use":        "docsctl-publish",
		"package":          packageName,
		"repository":       "go-go-golems/" + packageName,
		"repository_id":    "123456",
		"workflow_ref":     "go-go-golems/" + packageName + "/.github/workflows/publish-docs.yml@refs/heads/main",
		"job_workflow_ref": "go-go-golems/infra-tooling/.github/workflows/publish-docsctl.yml@refs/heads/main",
		"run_id":           "987654321",
	}
}

func TestJWTPublisherAuthAllowsMatchingPackage(t *testing.T) {
	issuer := newTestOIDCIssuer(t)
	auth := newTestJWTAuth(t, issuer)
	token := issuer.sign(t, validDocsPublishClaims("glazed"))

	identity, err := auth.AuthorizePublish(context.Background(), token, PublishRequest{PackageName: "glazed", Version: "vtest"})
	if err != nil {
		t.Fatalf("authorize publish: %v", err)
	}
	if identity.Method != "vault-oidc-jwt" {
		t.Fatalf("method = %q", identity.Method)
	}
	if identity.PackageName != "glazed" {
		t.Fatalf("package = %q", identity.PackageName)
	}
	if identity.Subject != "go-go-golems/glazed" {
		t.Fatalf("subject = %q", identity.Subject)
	}
}

func TestJWTPublisherAuthRejectsPackageMismatch(t *testing.T) {
	issuer := newTestOIDCIssuer(t)
	auth := newTestJWTAuth(t, issuer)
	token := issuer.sign(t, validDocsPublishClaims("glazed"))

	_, err := auth.AuthorizePublish(context.Background(), token, PublishRequest{PackageName: "pinocchio", Version: "vtest"})
	if err != ErrForbidden {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

func TestJWTPublisherAuthRejectsWrongAudience(t *testing.T) {
	issuer := newTestOIDCIssuer(t)
	auth := newTestJWTAuth(t, issuer)
	claims := validDocsPublishClaims("glazed")
	claims["aud"] = "other-service"
	token := issuer.sign(t, claims)

	_, err := auth.AuthorizePublish(context.Background(), token, PublishRequest{PackageName: "glazed", Version: "vtest"})
	if err != ErrUnauthorized {
		t.Fatalf("err = %v, want ErrUnauthorized", err)
	}
}

func TestJWTPublisherAuthRejectsWrongIssuer(t *testing.T) {
	issuer := newTestOIDCIssuer(t)
	auth := newTestJWTAuth(t, issuer)
	claims := validDocsPublishClaims("glazed")
	claims["iss"] = "https://vault.example.invalid/v1/identity/oidc"
	token := issuer.sign(t, claims)

	_, err := auth.AuthorizePublish(context.Background(), token, PublishRequest{PackageName: "glazed", Version: "vtest"})
	if err != ErrUnauthorized {
		t.Fatalf("err = %v, want ErrUnauthorized", err)
	}
}

func TestJWTPublisherAuthRejectsExpiredToken(t *testing.T) {
	issuer := newTestOIDCIssuer(t)
	auth := newTestJWTAuth(t, issuer)
	claims := validDocsPublishClaims("glazed")
	claims["exp"] = time.Now().Add(-time.Minute).Unix()
	token := issuer.sign(t, claims)

	_, err := auth.AuthorizePublish(context.Background(), token, PublishRequest{PackageName: "glazed", Version: "vtest"})
	if err != ErrUnauthorized {
		t.Fatalf("err = %v, want ErrUnauthorized", err)
	}
}

func TestJWTPublisherAuthRejectsTamperedToken(t *testing.T) {
	issuer := newTestOIDCIssuer(t)
	auth := newTestJWTAuth(t, issuer)
	token := issuer.sign(t, validDocsPublishClaims("glazed"))
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("unexpected token parts: %d", len(parts))
	}
	claims := validDocsPublishClaims("pinocchio")
	claims["iss"] = issuer.issuer
	claims["sub"] = "vault-entity-123"
	claims["aud"] = "docs-registry"
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(5 * time.Minute).Unix()
	tampered := parts[0] + "." + mustBase64JSON(t, claims) + "." + parts[2]

	_, err := auth.AuthorizePublish(context.Background(), tampered, PublishRequest{PackageName: "pinocchio", Version: "vtest"})
	if err != ErrUnauthorized {
		t.Fatalf("err = %v, want ErrUnauthorized", err)
	}
}

func TestJWTPublisherAuthRejectsMissingTokenUse(t *testing.T) {
	issuer := newTestOIDCIssuer(t)
	auth := newTestJWTAuth(t, issuer)
	claims := validDocsPublishClaims("glazed")
	delete(claims, "token_use")
	token := issuer.sign(t, claims)

	_, err := auth.AuthorizePublish(context.Background(), token, PublishRequest{PackageName: "glazed", Version: "vtest"})
	if err != ErrForbidden {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

func TestJWTPublisherAuthRejectsMissingPackage(t *testing.T) {
	issuer := newTestOIDCIssuer(t)
	auth := newTestJWTAuth(t, issuer)
	claims := validDocsPublishClaims("glazed")
	delete(claims, "package")
	token := issuer.sign(t, claims)

	_, err := auth.AuthorizePublish(context.Background(), token, PublishRequest{PackageName: "glazed", Version: "vtest"})
	if err != ErrForbidden {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}
