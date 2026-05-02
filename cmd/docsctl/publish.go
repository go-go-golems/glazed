package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/help/publish"
)

type PublishCommand struct {
	*cmds.CommandDescription
}

type publishOptions struct {
	Server      string `glazed:"server"`
	PackageName string `glazed:"package"`
	Version     string `glazed:"version"`
	File        string `glazed:"file"`
	Token       string `glazed:"token"`
	TokenFile   string `glazed:"token-file"`
	JSONOutput  bool   `glazed:"json"`
	DryRun      bool   `glazed:"dry-run"`
}

var _ cmds.WriterCommand = (*PublishCommand)(nil)

func NewPublishCommand() (*PublishCommand, error) {
	return &PublishCommand{CommandDescription: cmds.NewCommandDescription(
		"publish",
		cmds.WithShort("Publish a Glazed help SQLite database"),
		cmds.WithLong(`Publish validates a Glazed help SQLite export locally, then uploads it to
a docs registry using a package-scoped bearer token.

Token precedence is --token, DOCS_YOLO_PUBLISH_TOKEN, then --token-file.`),
		cmds.WithFlags(
			fields.New("server", fields.TypeString, fields.WithHelp("Docs registry base URL"), fields.WithRequired(true)),
			fields.New("package", fields.TypeString, fields.WithHelp("Package name"), fields.WithRequired(true)),
			fields.New("version", fields.TypeString, fields.WithHelp("Package version"), fields.WithRequired(true)),
			fields.New("file", fields.TypeString, fields.WithHelp("SQLite help export file"), fields.WithRequired(true)),
			fields.New("token", fields.TypeString, fields.WithHelp("Publish token (or DOCS_YOLO_PUBLISH_TOKEN)"), fields.WithDefault("")),
			fields.New("token-file", fields.TypeString, fields.WithHelp("File containing publish token"), fields.WithDefault("")),
			fields.New("json", fields.TypeBool, fields.WithHelp("Emit JSON"), fields.WithDefault(false)),
			fields.New("dry-run", fields.TypeBool, fields.WithHelp("Validate locally and print target without uploading"), fields.WithDefault(false)),
		),
	)}, nil
}

func (c *PublishCommand) RunIntoWriter(ctx context.Context, parsedValues *values.Values, w io.Writer) error {
	opts := &publishOptions{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, opts); err != nil {
		return err
	}
	return runPublish(ctx, w, opts)
}

func runPublish(ctx context.Context, w io.Writer, opts *publishOptions) error {
	result, err := publish.ValidateSQLiteHelpDB(ctx, opts.File, publish.SQLiteValidationOptions{PackageName: opts.PackageName, Version: opts.Version})
	if err != nil {
		return err
	}
	url := strings.TrimRight(opts.Server, "/") + fmt.Sprintf("/v1/packages/%s/versions/%s/sqlite", opts.PackageName, opts.Version)
	if opts.DryRun {
		payload := map[string]any{"ok": true, "dryRun": true, "url": url, "validation": result}
		return writePublishResult(w, opts.JSONOutput, payload, "DRY: %s is valid; would upload to %s\n", result.Path, url)
	}
	token, err := resolvePublishToken(opts)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(opts.File)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/vnd.sqlite3")
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("publish failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	if opts.JSONOutput {
		_, err = w.Write(body)
		return err
	}
	_, err = fmt.Fprintf(w, "OK: published %s@%s to %s\n", opts.PackageName, opts.Version, opts.Server)
	return err
}

func resolvePublishToken(opts *publishOptions) (string, error) {
	if opts.Token != "" {
		return opts.Token, nil
	}
	if env := os.Getenv("DOCS_YOLO_PUBLISH_TOKEN"); env != "" {
		return env, nil
	}
	if opts.TokenFile != "" {
		data, err := os.ReadFile(opts.TokenFile)
		if err != nil {
			return "", err
		}
		if token := strings.TrimSpace(string(data)); token != "" {
			return token, nil
		}
	}
	return "", fmt.Errorf("publish token is required via --token, DOCS_YOLO_PUBLISH_TOKEN, or --token-file")
}

func writePublishResult(w io.Writer, jsonOutput bool, payload any, format string, args ...any) error {
	if jsonOutput {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(payload)
	}
	_, err := fmt.Fprintf(w, format, args...)
	return err
}
