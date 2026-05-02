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

	"github.com/go-go-golems/glazed/pkg/help/publish"
	"github.com/spf13/cobra"
)

type publishOptions struct {
	server      string
	packageName string
	version     string
	file        string
	token       string
	tokenFile   string
	jsonOutput  bool
	dryRun      bool
}

func newPublishCommand() *cobra.Command {
	opts := &publishOptions{}
	cmd := &cobra.Command{Use: "publish --server <url> --package <name> --version <version> --file <help.db>", Short: "Publish a Glazed help SQLite database", RunE: func(cmd *cobra.Command, args []string) error { return runPublish(cmd, opts) }}
	cmd.Flags().StringVar(&opts.server, "server", "", "Docs registry base URL")
	cmd.Flags().StringVar(&opts.packageName, "package", "", "Package name")
	cmd.Flags().StringVar(&opts.version, "version", "", "Package version")
	cmd.Flags().StringVar(&opts.file, "file", "", "SQLite help export file")
	cmd.Flags().StringVar(&opts.token, "token", "", "Publish token (or DOCS_YOLO_PUBLISH_TOKEN)")
	cmd.Flags().StringVar(&opts.tokenFile, "token-file", "", "File containing publish token")
	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Emit JSON")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Validate locally and print target without uploading")
	_ = cmd.MarkFlagRequired("server")
	_ = cmd.MarkFlagRequired("package")
	_ = cmd.MarkFlagRequired("version")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func runPublish(cmd *cobra.Command, opts *publishOptions) error {
	result, err := publish.ValidateSQLiteHelpDB(context.Background(), opts.file, publish.SQLiteValidationOptions{PackageName: opts.packageName, Version: opts.version})
	if err != nil {
		return err
	}
	url := strings.TrimRight(opts.server, "/") + fmt.Sprintf("/v1/packages/%s/versions/%s/sqlite", opts.packageName, opts.version)
	if opts.dryRun {
		payload := map[string]any{"ok": true, "dryRun": true, "url": url, "validation": result}
		return writePublishResult(cmd, opts.jsonOutput, payload, "DRY: %s is valid; would upload to %s\n", result.Path, url)
	}
	token, err := resolvePublishToken(opts)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(opts.file)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, url, bytes.NewReader(data))
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
	if opts.jsonOutput {
		_, err = cmd.OutOrStdout().Write(body)
		return err
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "OK: published %s@%s to %s\n", opts.packageName, opts.version, opts.server)
	return err
}

func resolvePublishToken(opts *publishOptions) (string, error) {
	if opts.token != "" {
		return opts.token, nil
	}
	if env := os.Getenv("DOCS_YOLO_PUBLISH_TOKEN"); env != "" {
		return env, nil
	}
	if opts.tokenFile != "" {
		data, err := os.ReadFile(opts.tokenFile)
		if err != nil {
			return "", err
		}
		if token := strings.TrimSpace(string(data)); token != "" {
			return token, nil
		}
	}
	return "", fmt.Errorf("publish token is required via --token, DOCS_YOLO_PUBLISH_TOKEN, or --token-file")
}

func writePublishResult(cmd *cobra.Command, jsonOutput bool, payload any, format string, args ...any) error {
	if jsonOutput {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(payload)
	}
	_, err := fmt.Fprintf(cmd.OutOrStdout(), format, args...)
	return err
}
