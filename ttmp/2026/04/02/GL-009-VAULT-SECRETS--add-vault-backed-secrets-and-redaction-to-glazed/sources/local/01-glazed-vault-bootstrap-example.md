---
Title: Glazed vault bootstrap example
Ticket: GL-009-VAULT-SECRETS
Status: active
Topics:
    - glazed
    - config
    - vault
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources:
    - local:glazed-vault-bootstrap-example.md
Summary: Imported source artifact preserving the bootstrap-parse sketch provided by the user.
LastUpdated: 2026-04-02T19:20:42.706701963-04:00
WhatFor: Imported proposal artifact
WhenToUse: Compare the original sketch against the final ticket recommendation.
---

# Glazed secret loading pattern

This is the split that keeps the design clean.

## 1) Keep the resolver middleware pure

`FromVaultSettings` should assume the Vault/KMS settings are already resolved.
It just overlays secret values onto sensitive fields.

```go
func FromVaultSettings(vs *VaultSettings, options ...fields.ParseOption) sources.Middleware {
    return func(next sources.HandlerFunc) sources.HandlerFunc {
        return func(schema_ *schema.Schema, parsedValues *values.Values) error {
            if err := next(schema_, parsedValues); err != nil {
                return err
            }

            ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
            defer cancel()

            client, err := NewVaultClient(ctx, vs)
            if err != nil {
                return err
            }

            secrets, err := client.ReadPath(vs.SecretPath)
            if err != nil {
                return err
            }

            return schema_.ForEachE(func(_ string, section schema.Section) error {
                sectionValues := parsedValues.GetOrCreate(section)
                return section.GetDefinitions().ForEachE(func(def *fields.Definition) error {
                    // Recommended: only hydrate secret/sensitive fields.
                    if def.Type != fields.TypeSecret {
                        return nil
                    }

                    raw, ok := secrets[def.Name]
                    if !ok {
                        return nil
                    }

                    opts := append([]fields.ParseOption{
                        fields.WithSource("vault"),
                        fields.WithMetadata(map[string]interface{}{
                            "provider": "vault",
                            "path":     vs.SecretPath,
                        }),
                    }, options...)

                    return sectionValues.Fields.UpdateValue(def.Name, def, raw, opts...)
                })
            })
        }
    }
}
```

## 2) Bootstrap-parse the Vault settings section only when needed

You need this when Vault settings may come from env/flags/config, but env/flags should still be able to override the _final_ secret-backed fields.

```go
func buildMiddlewares(
    cmd *cobra.Command,
    args []string,
    configFiles []string,
    vaultSettingsSection schema.Section,
) ([]sources.Middleware, error) {
    // Bootstrap parse: resolve only vault-settings using the sources that are
    // allowed to influence provider configuration.
    bootstrapSchema := schema.NewSchema(schema.WithSections(vaultSettingsSection))
    bootstrapValues := values.New()

    if err := sources.Execute(
        bootstrapSchema,
        bootstrapValues,
        sources.FromCobra(cmd, fields.WithSource("cobra")),
        sources.FromEnv("MYAPP", fields.WithSource("env")),
        sources.FromFiles(configFiles, sources.WithParseOptions(fields.WithSource("config"))),
        sources.FromDefaults(fields.WithSource(fields.SourceDefaults)),
    ); err != nil {
        return nil, err
    }

    vs := &VaultSettings{}
    if err := bootstrapValues.DecodeSectionInto("vault-settings", vs); err != nil {
        return nil, err
    }

    // Main chain: place Vault between config and env/cobra so env/cobra still win.
    return []sources.Middleware{
        sources.FromCobra(cmd, fields.WithSource("cobra")),
        sources.FromEnv("MYAPP", fields.WithSource("env")),
        FromVaultSettings(vs),
        sources.FromFiles(configFiles, sources.WithParseOptions(fields.WithSource("config"))),
        sources.FromDefaults(fields.WithSource(fields.SourceDefaults)),
    }, nil
}
```

## 3) Why this split matters

Without bootstrap parsing, you cannot have both of these at the same time:

- env/flags influence the Vault client settings
- env/flags still override the final secret-populated application fields

A single pass forces env/flags to be either before Vault or after Vault, but not both.

## 4) Field selection

For a first pass, use `def.Type == fields.TypeSecret` as the selector.
That is safer than your old “same name exists in Vault, so overwrite it” rule.

If you later need non-1:1 mappings, add a dedicated field option like `SecretRef`.
