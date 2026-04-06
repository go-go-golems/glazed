package sources

import (
	"context"
	stdErrors "errors"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

type fakeVaultClient struct {
	readPath       string
	secrets        map[string]interface{}
	templateCtx    vaultTemplateContext
	buildCtxCalled bool
}

func (f *fakeVaultClient) ReadPath(_ context.Context, path string) (map[string]interface{}, error) {
	f.readPath = path
	return f.secrets, nil
}

func (f *fakeVaultClient) BuildTemplateContext(_ context.Context) (vaultTemplateContext, error) {
	f.buildCtxCalled = true
	return f.templateCtx, nil
}

type roundTripFunc func(*http.Request) (*http.Response, error)

var errContextNotPropagated = stdErrors.New("request context was not canceled")

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newTestAPIVaultClient(t *testing.T, rt http.RoundTripper) *apiVaultClient {
	t.Helper()

	config := api.DefaultConfig()
	config.Address = "http://vault.test"
	config.HttpClient = &http.Client{Transport: rt}
	config.MaxRetries = 0
	config.Timeout = 0

	client, err := api.NewClient(config)
	require.NoError(t, err)

	return &apiVaultClient{client: client}
}

func TestGetVaultSettingsDecodesSection(t *testing.T) {
	section, err := NewVaultSettingsSection()
	require.NoError(t, err)

	parsed := values.New()
	schema_ := schema.NewSchema(schema.WithSections(section))
	err = Execute(
		schema_,
		parsed,
		FromMap(map[string]map[string]interface{}{
			VaultSettingsSlug: {
				"vault-addr":         "https://vault.example.com",
				"vault-token":        "token-123",
				"vault-token-source": string(VaultTokenSourceFile),
				"vault-token-file":   "/tmp/token",
				"secret-path":        "kv/my-app",
			},
		}),
		FromDefaults(fields.WithSource(fields.SourceDefaults)),
	)
	require.NoError(t, err)

	settings, err := GetVaultSettings(parsed)
	require.NoError(t, err)
	require.Equal(t, "https://vault.example.com", settings.VaultAddr)
	require.Equal(t, "token-123", settings.VaultToken)
	require.Equal(t, string(VaultTokenSourceFile), settings.VaultTokenSource)
	require.Equal(t, "/tmp/token", settings.VaultTokenFile)
	require.Equal(t, "kv/my-app", settings.SecretPath)
}

func TestFromVaultSettingsOnlyHydratesSensitiveFields(t *testing.T) {
	appSection, err := schema.NewSection(
		"app",
		"App",
		schema.WithFields(
			fields.New("username", fields.TypeString, fields.WithDefault("from-default-user")),
			fields.New("password", fields.TypeSecret, fields.WithDefault("from-default-password")),
		),
	)
	require.NoError(t, err)

	parsed := values.New()
	schema_ := schema.NewSchema(schema.WithSections(appSection))
	client := &fakeVaultClient{
		secrets: map[string]interface{}{
			"username": "from-vault-user",
			"password": "from-vault-password",
		},
	}

	err = Execute(
		schema_,
		parsed,
		fromVaultSettingsWithClientFactory(
			&VaultSettings{SecretPath: "kv/app"},
			func(_ context.Context, _ *VaultSettings) (vaultClient, error) {
				return client, nil
			},
		),
		FromMap(map[string]map[string]interface{}{
			"app": {
				"username": "from-config-user",
				"password": "from-config-password",
			},
		}),
		FromDefaults(fields.WithSource(fields.SourceDefaults)),
	)
	require.NoError(t, err)

	username, ok := parsed.GetField("app", "username")
	require.True(t, ok)
	require.Equal(t, "from-config-user", username.Value)

	password, ok := parsed.GetField("app", "password")
	require.True(t, ok)
	require.Equal(t, "from-vault-password", password.Value)
	require.Equal(t, "vault", password.Log[len(password.Log)-1].Source)
	require.Equal(t, "vault", password.Log[len(password.Log)-1].Metadata["provider"])
	require.Equal(t, "kv/app", password.Log[len(password.Log)-1].Metadata["path"])
}

func TestFromVaultSettingsRendersTemplatedPath(t *testing.T) {
	appSection, err := schema.NewSection(
		"app",
		"App",
		schema.WithFields(
			fields.New("password", fields.TypeSecret),
		),
	)
	require.NoError(t, err)

	parsed := values.New()
	schema_ := schema.NewSchema(schema.WithSections(appSection))
	client := &fakeVaultClient{
		secrets: map[string]interface{}{
			"password": "from-vault",
		},
		templateCtx: vaultTemplateContext{
			Token: vaultTokenContext{
				OIDCUserID: "user-42",
			},
		},
	}

	err = Execute(
		schema_,
		parsed,
		fromVaultSettingsWithClientFactory(
			&VaultSettings{SecretPath: "kv/apps/{{ .Token.OIDCUserID }}"},
			func(_ context.Context, _ *VaultSettings) (vaultClient, error) {
				return client, nil
			},
		),
	)
	require.NoError(t, err)
	require.True(t, client.buildCtxCalled)
	require.Equal(t, "kv/apps/user-42", client.readPath)
}

func TestBootstrapVaultSettingsPrecedence(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(configFile, []byte(`vault-settings:
  vault-addr: https://vault.from-config
  secret-path: kv/from-config
`), 0o644))

	t.Setenv("MYAPP_VAULT_ADDR", "https://vault.from-env")
	t.Setenv("MYAPP_SECRET_PATH", "kv/from-env")

	rootCmd := &cobra.Command{Use: "test"}
	section, err := NewVaultSettingsSection()
	require.NoError(t, err)
	require.NoError(t, section.(schema.CobraSection).AddSectionToCobraCommand(rootCmd))
	rootCmd.SetArgs([]string{"--secret-path", "kv/from-flag"})
	require.NoError(t, rootCmd.Execute())

	settings, err := BootstrapVaultSettings([]string{configFile}, []string{"MYAPP"}, rootCmd)
	require.NoError(t, err)
	require.Equal(t, "https://vault.from-env", settings.VaultAddr)
	require.Equal(t, "kv/from-flag", settings.SecretPath)
}

func TestBootstrapVaultSettingsMainChainPrecedence(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(configFile, []byte(`vault-settings:
  secret-path: kv/from-config
app:
  password: from-config
  host: host-from-config
`), 0o644))

	t.Setenv("MYAPP_PASSWORD", "from-env")

	appSection, err := schema.NewSection(
		"app",
		"App",
		schema.WithFields(
			fields.New("host", fields.TypeString, fields.WithDefault("host-from-default")),
			fields.New("password", fields.TypeSecret, fields.WithDefault("from-default")),
		),
	)
	require.NoError(t, err)

	rootCmd := &cobra.Command{Use: "test"}
	vaultSection, err := NewVaultSettingsSection()
	require.NoError(t, err)
	require.NoError(t, vaultSection.(schema.CobraSection).AddSectionToCobraCommand(rootCmd))
	require.NoError(t, appSection.AddSectionToCobraCommand(rootCmd))
	rootCmd.SetArgs([]string{"--password", "from-flag"})
	require.NoError(t, rootCmd.Execute())

	bootstrapSettings, err := BootstrapVaultSettings([]string{configFile}, []string{"MYAPP"}, rootCmd)
	require.NoError(t, err)
	require.Equal(t, "kv/from-config", bootstrapSettings.SecretPath)

	client := &fakeVaultClient{
		secrets: map[string]interface{}{
			"password": "from-vault",
		},
	}
	parsed := values.New()
	schema_ := schema.NewSchema(schema.WithSections(appSection))

	err = Execute(
		schema_,
		parsed,
		FromCobra(rootCmd, fields.WithSource("cobra")),
		FromEnv("MYAPP", fields.WithSource("env")),
		fromVaultSettingsWithClientFactory(
			bootstrapSettings,
			func(_ context.Context, _ *VaultSettings) (vaultClient, error) {
				return client, nil
			},
		),
		FromFiles([]string{configFile}, WithParseOptions(fields.WithSource("config"))),
		FromDefaults(fields.WithSource(fields.SourceDefaults)),
	)
	require.NoError(t, err)

	host, ok := parsed.GetField("app", "host")
	require.True(t, ok)
	require.Equal(t, "host-from-config", host.Value)

	password, ok := parsed.GetField("app", "password")
	require.True(t, ok)
	require.Equal(t, "from-flag", password.Value)
	require.Equal(t, "cobra", password.Log[len(password.Log)-1].Source)
}

func TestResolveVaultTokenAutoUsesExplicitTokenBeforeEnvironment(t *testing.T) {
	t.Setenv("VAULT_TOKEN", "from-env")

	token, err := resolveVaultToken(context.Background(), "from-explicit", VaultTokenSourceAuto, "")
	require.NoError(t, err)
	require.Equal(t, "from-explicit", token)
}

func TestResolveVaultTokenFileExpandsHomeDirectory(t *testing.T) {
	dir := t.TempDir()
	homeDir := filepath.Join(dir, "home")
	require.NoError(t, os.MkdirAll(homeDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, ".vault-token"), []byte("from-file\n"), 0o644))
	t.Setenv("HOME", homeDir)

	token, err := resolveVaultToken(context.Background(), "", VaultTokenSourceFile, "~/.vault-token")
	require.NoError(t, err)
	require.Equal(t, "from-file", token)
}

func TestResolveVaultTokenFileRejectsEmptyContent(t *testing.T) {
	dir := t.TempDir()
	tokenFile := filepath.Join(dir, "empty-token")
	require.NoError(t, os.WriteFile(tokenFile, []byte(" \n\t "), 0o644))

	token, err := resolveVaultToken(context.Background(), "", VaultTokenSourceFile, tokenFile)
	require.Empty(t, token)
	require.Error(t, err)
	require.ErrorContains(t, err, "is empty")
}

func TestAPIVaultClientReadPathHonorsContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestAPIVaultClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Context().Err() == nil {
			return nil, errContextNotPropagated
		}
		return nil, req.Context().Err()
	}))

	_, err := client.ReadPath(ctx, "secret/demo")
	require.Error(t, err)
	require.ErrorContains(t, err, context.Canceled.Error())
	require.NotErrorIs(t, err, errContextNotPropagated)
}

func TestAPIVaultClientBuildTemplateContextHonorsContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestAPIVaultClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Context().Err() == nil {
			return nil, errContextNotPropagated
		}
		return nil, req.Context().Err()
	}))

	_, err := client.BuildTemplateContext(ctx)
	require.Error(t, err)
	require.ErrorContains(t, err, context.Canceled.Error())
	require.NotErrorIs(t, err, errContextNotPropagated)
}
