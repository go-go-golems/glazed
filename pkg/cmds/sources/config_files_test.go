package sources_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/helpers"
	"github.com/go-go-golems/glazed/pkg/cmds/sources"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	glazedconfig "github.com/go-go-golems/glazed/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestFromFilesAddsConfigMetadata(t *testing.T) {
	tmp := t.TempDir()
	configFile := filepath.Join(tmp, "config.yaml")
	require.NoError(t, os.WriteFile(configFile, []byte("app:\n  host: from-config\n"), 0o644))

	schema_ := helpers.NewTestSchema([]helpers.TestSection{{
		Name: "app",
		Definitions: []*fields.Definition{
			fields.New("host", fields.TypeString),
		},
	}})
	parsed := values.New()

	err := sources.Execute(
		schema_,
		parsed,
		sources.FromFiles([]string{configFile}),
	)
	require.NoError(t, err)

	host, ok := parsed.GetField("app", "host")
	require.True(t, ok)
	require.Equal(t, "from-config", host.Value)
	require.NotEmpty(t, host.Log)

	step := host.Log[len(host.Log)-1]
	require.Equal(t, "config", step.Source)
	require.Equal(t, configFile, step.Metadata["config_file"])
	require.Equal(t, 0, step.Metadata["index"])
	require.Equal(t, 0, step.Metadata["config_index"])
	require.Equal(t, "files", step.Metadata["config_source_name"])
	require.Equal(t, "config-file", step.Metadata["config_source_kind"])
}

func TestFromResolvedFilesAddsLayeredConfigMetadata(t *testing.T) {
	tmp := t.TempDir()
	configFile := filepath.Join(tmp, ".pinocchio-profile.yml")
	require.NoError(t, os.WriteFile(configFile, []byte("app:\n  host: from-repo\n"), 0o644))

	schema_ := helpers.NewTestSchema([]helpers.TestSection{{
		Name: "app",
		Definitions: []*fields.Definition{
			fields.New("host", fields.TypeString),
		},
	}})
	parsed := values.New()

	err := sources.Execute(
		schema_,
		parsed,
		sources.FromResolvedFiles([]glazedconfig.ResolvedConfigFile{{
			Path:       configFile,
			Layer:      glazedconfig.LayerRepo,
			SourceName: "git-root-local-profile",
			SourceKind: "profile-overlay",
			Index:      3,
		}}),
	)
	require.NoError(t, err)

	host, ok := parsed.GetField("app", "host")
	require.True(t, ok)
	require.Equal(t, "from-repo", host.Value)
	require.NotEmpty(t, host.Log)

	step := host.Log[len(host.Log)-1]
	require.Equal(t, "config", step.Source)
	require.Equal(t, configFile, step.Metadata["config_file"])
	require.Equal(t, 3, step.Metadata["index"])
	require.Equal(t, 3, step.Metadata["config_index"])
	require.Equal(t, "repo", step.Metadata["config_layer"])
	require.Equal(t, "git-root-local-profile", step.Metadata["config_source_name"])
	require.Equal(t, "profile-overlay", step.Metadata["config_source_kind"])
}
