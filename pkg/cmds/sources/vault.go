package sources

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type vaultClient interface {
	ReadPath(ctx context.Context, path string) (map[string]interface{}, error)
	BuildTemplateContext(ctx context.Context) (vaultTemplateContext, error)
}

type vaultClientFactory func(ctx context.Context, settings *VaultSettings) (vaultClient, error)

type apiVaultClient struct {
	client *api.Client
}

type vaultTemplateContext struct {
	Token vaultTokenContext
	Extra map[string]interface{}
	Data  map[string]interface{}
}

type vaultTokenContext struct {
	Accessor    string
	CreationTTL string
	DisplayName string
	EntityID    string
	ExpireTime  string
	ID          string
	IssueTime   string
	Meta        map[string]string
	Policies    []string
	OIDCUserID  string
	Path        string
	TTL         string
	Type        string
}

func BootstrapVaultSettings(
	configFiles []string,
	envPrefixes []string,
	cmd *cobra.Command,
) (*VaultSettings, error) {
	section, err := NewVaultSettingsSection()
	if err != nil {
		return nil, err
	}

	bootstrapSchema := schema.NewSchema(schema.WithSections(section))
	bootstrapValues := values.New()
	bootstrapMiddlewares := []Middleware{}

	if cmd != nil {
		bootstrapMiddlewares = append(bootstrapMiddlewares,
			FromCobra(cmd, fields.WithSource("cobra")),
		)
	}

	for i := len(envPrefixes) - 1; i >= 0; i-- {
		prefix := strings.TrimSpace(envPrefixes[i])
		if prefix == "" {
			continue
		}
		bootstrapMiddlewares = append(bootstrapMiddlewares,
			FromEnv(prefix, fields.WithSource("env")),
		)
	}

	if len(configFiles) > 0 {
		bootstrapMiddlewares = append(bootstrapMiddlewares,
			FromFiles(
				configFiles,
				WithParseOptions(fields.WithSource("config")),
			),
		)
	}

	bootstrapMiddlewares = append(bootstrapMiddlewares,
		FromDefaults(fields.WithSource(fields.SourceDefaults)),
	)

	if err := Execute(bootstrapSchema, bootstrapValues, bootstrapMiddlewares...); err != nil {
		return nil, errors.Wrap(err, "failed to bootstrap-parse vault-settings")
	}

	settings, err := GetVaultSettings(bootstrapValues)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize bootstrap vault settings")
	}

	return settings, nil
}

func FromVaultSettings(vs *VaultSettings, options ...fields.ParseOption) Middleware {
	return fromVaultSettingsWithClientFactory(vs, newVaultClientFromSettings, options...)
}

func fromVaultSettingsWithClientFactory(
	vs *VaultSettings,
	factory vaultClientFactory,
	options ...fields.ParseOption,
) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			if err := next(schema_, parsedValues); err != nil {
				return err
			}

			if vs == nil || strings.TrimSpace(vs.SecretPath) == "" {
				return nil
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			client, err := factory(ctx, vs)
			if err != nil {
				return errors.Wrap(err, "failed to initialize vault client")
			}

			effectivePath, err := renderVaultPath(ctx, client, vs.SecretPath)
			if err != nil {
				return errors.Wrap(err, "failed to resolve vault secret path")
			}

			secrets, err := client.ReadPath(ctx, effectivePath)
			if err != nil {
				return errors.Wrapf(err, "failed to read vault secrets from %s", effectivePath)
			}

			return schema_.ForEachE(func(_ string, section schema.Section) error {
				if section.GetSlug() == VaultSettingsSlug {
					return nil
				}

				sectionValues := parsedValues.GetOrCreate(section)
				return section.GetDefinitions().ForEachE(func(definition *fields.Definition) error {
					if !definition.Type.IsSensitive() {
						return nil
					}

					rawValue, ok := secrets[definition.Name]
					if !ok {
						return nil
					}

					parseOptions := append([]fields.ParseOption{
						fields.WithSource("vault"),
						fields.WithMetadata(map[string]interface{}{
							"provider": "vault",
							"path":     effectivePath,
						}),
					}, options...)

					if err := sectionValues.Fields.UpdateValue(definition.Name, definition, rawValue, parseOptions...); err != nil {
						return errors.Wrapf(err, "failed to apply vault secret to %s.%s", section.GetSlug(), definition.Name)
					}

					return nil
				})
			})
		}
	}
}

func newVaultClientFromSettings(ctx context.Context, settings *VaultSettings) (vaultClient, error) {
	if settings == nil {
		return nil, errors.New("vault settings must not be nil")
	}

	address := strings.TrimSpace(settings.VaultAddr)
	if address == "" {
		address = "http://127.0.0.1:8200"
	}

	token, err := resolveVaultToken(
		ctx,
		settings.VaultToken,
		VaultTokenSource(settings.VaultTokenSource),
		settings.VaultTokenFile,
	)
	if err != nil {
		return nil, err
	}

	config := api.DefaultConfig()
	config.Address = address

	client, err := api.NewClient(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Vault client")
	}
	client.SetToken(token)

	return &apiVaultClient{client: client}, nil
}

func (c *apiVaultClient) ReadPath(ctx context.Context, path string) (map[string]interface{}, error) {
	effectivePath := strings.TrimSpace(path)
	if effectivePath == "" {
		return nil, errors.New("vault path is empty")
	}

	mountPath, secretPath := parseVaultPath(effectivePath)

	if data, err := c.readKVv2Path(ctx, mountPath, secretPath); err == nil {
		return data, nil
	}

	data, err := c.readKVv1Path(ctx, effectivePath)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (c *apiVaultClient) BuildTemplateContext(ctx context.Context) (vaultTemplateContext, error) {
	lookup, err := c.client.Auth().Token().LookupSelfWithContext(ctx)
	if err != nil {
		return vaultTemplateContext{}, errors.Wrap(err, "failed to lookup current Vault token")
	}

	ret := vaultTemplateContext{
		Token: vaultTokenContext{
			Meta: map[string]string{},
		},
	}

	if lookup == nil || lookup.Data == nil {
		return ret, nil
	}

	getString := func(key string) string {
		value, ok := lookup.Data[key]
		if !ok {
			return ""
		}

		switch v := value.(type) {
		case string:
			return v
		default:
			return fmt.Sprint(v)
		}
	}

	ret.Token.Accessor = getString("accessor")
	ret.Token.CreationTTL = getString("creation_ttl")
	ret.Token.DisplayName = getString("display_name")
	ret.Token.EntityID = getString("entity_id")
	ret.Token.ExpireTime = getString("expire_time")
	ret.Token.ID = getString("id")
	ret.Token.IssueTime = getString("issue_time")
	ret.Token.Path = getString("path")
	ret.Token.TTL = getString("ttl")
	ret.Token.Type = getString("type")

	if policies, ok := lookup.Data["policies"].([]interface{}); ok {
		for _, policy := range policies {
			policyString, ok := policy.(string)
			if ok {
				ret.Token.Policies = append(ret.Token.Policies, policyString)
			}
		}
	}

	if metadata, ok := lookup.Data["meta"].(map[string]interface{}); ok {
		for key, value := range metadata {
			valueString, ok := value.(string)
			if ok {
				ret.Token.Meta[key] = valueString
			}
		}
	}

	if strings.HasPrefix(ret.Token.DisplayName, "oidc-") {
		re := regexp.MustCompile(`oidc-([0-9A-Za-z_-]+)`)
		match := re.FindStringSubmatch(ret.Token.DisplayName)
		if len(match) == 2 {
			ret.Token.OIDCUserID = match[1]
		}
	}

	return ret, nil
}

func (c *apiVaultClient) readKVv1Path(ctx context.Context, path string) (map[string]interface{}, error) {
	secret, err := c.client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read Vault secret from %s", path)
	}
	if secret == nil {
		return nil, errors.Errorf("no Vault secret found at %s", path)
	}
	return secret.Data, nil
}

func (c *apiVaultClient) readKVv2Path(ctx context.Context, mountPath string, secretPath string) (map[string]interface{}, error) {
	fullPath := fmt.Sprintf("%s/data/%s", mountPath, secretPath)
	secret, err := c.client.Logical().ReadWithContext(ctx, fullPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read Vault KV v2 secret from %s", fullPath)
	}
	if secret == nil {
		return nil, errors.Errorf("no Vault KV v2 secret found at %s", fullPath)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, errors.Errorf("invalid Vault KV v2 secret format at %s", fullPath)
	}

	return data, nil
}

func renderVaultPath(ctx context.Context, client vaultClient, path string) (string, error) {
	effectivePath := strings.TrimSpace(path)
	if effectivePath == "" {
		return "", errors.New("vault path is empty")
	}

	if !strings.Contains(effectivePath, "{{") {
		return effectivePath, nil
	}

	templateContext, err := client.BuildTemplateContext(ctx)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("vault-path").Option("missingkey=error").Parse(effectivePath)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, templateContext); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func resolveVaultToken(
	ctx context.Context,
	explicitToken string,
	source VaultTokenSource,
	tokenFilePath string,
) (string, error) {
	if source == "" {
		source = VaultTokenSourceAuto
	}

	if strings.HasPrefix(tokenFilePath, "~") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			tokenFilePath = filepath.Join(homeDir, strings.TrimPrefix(tokenFilePath, "~"))
		}
	}

	switch source {
	case VaultTokenSourceEnv:
		if explicitToken != "" {
			return explicitToken, nil
		}
		token := os.Getenv("VAULT_TOKEN")
		if token == "" {
			return "", errors.New("no Vault token found in environment")
		}
		return token, nil

	case VaultTokenSourceFile:
		if tokenFilePath == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", errors.Wrap(err, "failed to resolve home directory for vault token file")
			}
			tokenFilePath = filepath.Join(homeDir, ".vault-token")
		}
		data, err := os.ReadFile(tokenFilePath)
		if err != nil {
			return "", errors.Wrapf(err, "failed to read Vault token file %s", tokenFilePath)
		}
		token := strings.TrimSpace(string(data))
		if token == "" {
			return "", errors.Errorf("Vault token file %s is empty", tokenFilePath)
		}
		return token, nil

	case VaultTokenSourceLookup:
		return lookupVaultToken(ctx)

	case VaultTokenSourceAuto:
		if explicitToken != "" {
			return explicitToken, nil
		}
		if token := os.Getenv("VAULT_TOKEN"); token != "" {
			return token, nil
		}
		if token, err := resolveVaultToken(ctx, "", VaultTokenSourceFile, tokenFilePath); err == nil && token != "" {
			return token, nil
		}
		if token, err := resolveVaultToken(ctx, "", VaultTokenSourceLookup, tokenFilePath); err == nil && token != "" {
			return token, nil
		}
		return "", errors.New("unable to resolve Vault token (tried explicit token, VAULT_TOKEN, token file, and vault token lookup)")

	default:
		return "", errors.Errorf("unknown Vault token source %q", source)
	}
}

func lookupVaultToken(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "vault", "token", "lookup", "-format=json")
	cmd.Env = os.Environ()

	output, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "failed to execute vault token lookup")
	}

	var payload struct {
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		return "", errors.Wrap(err, "failed to parse vault token lookup output")
	}

	value, ok := payload.Data["id"]
	if !ok {
		return "", errors.New("vault token lookup output missing data.id")
	}

	token, ok := value.(string)
	if !ok || token == "" {
		return "", errors.New("vault token lookup output did not contain a usable token id")
	}

	return token, nil
}

func parseVaultPath(path string) (string, string) {
	parts := strings.SplitN(strings.TrimSpace(path), "/", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}
