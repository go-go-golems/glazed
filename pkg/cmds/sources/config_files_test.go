package sources_test

import (
	"context"
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

func TestFromConfigPlanAddsLayeredConfigMetadata(t *testing.T) {
	tmp := t.TempDir()
	configFile := filepath.Join(tmp, ".pinocchio-profile.yml")
	require.NoError(t, os.WriteFile(configFile, []byte("app:\n  host: from-plan\n"), 0o644))

	schema_ := helpers.NewTestSchema([]helpers.TestSection{{
		Name: "app",
		Definitions: []*fields.Definition{
			fields.New("host", fields.TypeString),
		},
	}})
	parsed := values.New()

	plan := glazedconfig.NewPlan(
		glazedconfig.WithLayerOrder(glazedconfig.LayerRepo),
	).Add(
		glazedconfig.ExplicitFile(configFile).Named("repo-config").InLayer(glazedconfig.LayerRepo).Kind("profile-overlay"),
	)

	err := sources.Execute(
		schema_,
		parsed,
		sources.FromConfigPlan(plan),
	)
	require.NoError(t, err)

	host, ok := parsed.GetField("app", "host")
	require.True(t, ok)
	require.Equal(t, "from-plan", host.Value)
	require.NotEmpty(t, host.Log)

	step := host.Log[len(host.Log)-1]
	require.Equal(t, "config", step.Source)
	require.Equal(t, configFile, step.Metadata["config_file"])
	require.Equal(t, "repo", step.Metadata["config_layer"])
	require.Equal(t, "repo-config", step.Metadata["config_source_name"])
	require.Equal(t, "profile-overlay", step.Metadata["config_source_kind"])
}

func TestFromConfigPlanBuilderUsesParsedValues(t *testing.T) {
	tmp := t.TempDir()
	configFile := filepath.Join(tmp, "selected.yaml")
	require.NoError(t, os.WriteFile(configFile, []byte("app:\n  host: from-builder\n"), 0o644))

	schema_ := helpers.NewTestSchema([]helpers.TestSection{{
		Name: "app",
		Definitions: []*fields.Definition{
			fields.New("host", fields.TypeString),
		},
	}, {
		Name: "selector",
		Definitions: []*fields.Definition{
			fields.New("config-path", fields.TypeString),
		},
	}})
	parsed := values.New()

	err := sources.Execute(
		schema_,
		parsed,
		sources.FromConfigPlanBuilder(func(_ context.Context, parsedValues *values.Values) (*glazedconfig.Plan, error) {
			section := &struct {
				ConfigPath string `glazed:"config-path"`
			}{}
			if err := parsedValues.DecodeSectionInto("selector", section); err != nil {
				return nil, err
			}
			return glazedconfig.NewPlan(
				glazedconfig.WithLayerOrder(glazedconfig.LayerExplicit),
			).Add(
				glazedconfig.ExplicitFile(section.ConfigPath).Named("selected-config").Kind("explicit-file"),
			), nil
		}),
		sources.FromMap(map[string]map[string]interface{}{
			"selector": {"config-path": configFile},
		}),
	)
	require.NoError(t, err)

	host, ok := parsed.GetField("app", "host")
	require.True(t, ok)
	require.Equal(t, "from-builder", host.Value)
	require.Equal(t, "selected-config", host.Log[len(host.Log)-1].Metadata["config_source_name"])
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
