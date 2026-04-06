package fields

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestRenderValueSecretMasked(t *testing.T) {
	rendered, err := RenderValue(TypeSecret, "super-secret-token")
	require.NoError(t, err)
	require.Equal(t, "su***en", rendered)
}

func TestToSerializableFieldValueRedactsSensitiveData(t *testing.T) {
	definition := New("api-key", TypeSecret)
	fieldValue := &FieldValue{
		Definition: definition,
		Value:      "super-secret-token",
		Log: []ParseStep{{
			Source: "config",
			Value:  "super-secret-token",
			Metadata: map[string]interface{}{
				"map-value":      "super-secret-token",
				"parsed-strings": []string{"super-secret-token"},
				"index":          1,
			},
		}},
	}

	serializable := ToSerializableFieldValue(fieldValue)
	require.Equal(t, "su***en", serializable.Value)
	require.Len(t, serializable.Log, 1)
	require.Equal(t, "su***en", serializable.Log[0].Value)
	require.Equal(t, "su***en", serializable.Log[0].Metadata["map-value"])
	require.Equal(t, []string{"su***en"}, serializable.Log[0].Metadata["parsed-strings"])
	require.Equal(t, 1, serializable.Log[0].Metadata["index"])
}

func TestFieldValuesJSONMarshallingRedactsSensitiveData(t *testing.T) {
	definition := New("api-key", TypeSecret)
	values := NewFieldValues()
	require.NoError(t, values.UpdateValue("api-key", definition, "super-secret-token", WithSource("config")))

	data, err := json.Marshal(values)
	require.NoError(t, err)
	require.NotContains(t, string(data), "super-secret-token")
	require.Contains(t, string(data), "su***en")
}

func TestAddFieldsToCobraCommandRedactsSensitiveDefaults(t *testing.T) {
	defs := NewDefinitions(WithDefinitionList([]*Definition{
		New("api-key", TypeSecret, WithDefault("super-secret-token")),
	}))

	cmd := &cobra.Command{Use: "test"}
	require.NoError(t, defs.AddFieldsToCobraCommand(cmd, ""))

	flag := cmd.Flags().Lookup("api-key")
	require.NotNil(t, flag)
	require.Equal(t, "su***en", flag.DefValue)

	v, err := cmd.Flags().GetString("api-key")
	require.NoError(t, err)
	require.Equal(t, "super-secret-token", v)
}
