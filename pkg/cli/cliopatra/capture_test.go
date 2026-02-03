package cliopatra

import (
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeParsedDefaultLayer(desc *cmds.CommandDescription, ps *fields.FieldValues) *values.Values {
	defaultSection, ok := desc.GetLayer(schema.DefaultSlug)
	if !ok {
		return nil
	}

	ret := values.New()
	ret.Set(schema.DefaultSlug, &values.SectionValues{
		Section: defaultSection,
		Fields:  ps,
	})

	return ret
}

func TestSingleFlag(t *testing.T) {
	testPd := fields.New("test", fields.TypeString)
	desc := cmds.NewCommandDescription("test",
		cmds.WithFlags(
			testPd,
		),
	)
	p := NewProgramFromCapture(
		desc,
		makeParsedDefaultLayer(desc, fields.NewFieldValues(fields.WithFieldValue(testPd, "test", "foobar"))),
	)

	assert.Equal(t, "test", p.Name)
	assert.Equal(t, "", p.Description)
	assert.Len(t, p.Flags, 1)
	assert.Equal(t, "test", p.Flags[0].Name)
	assert.Equal(t, "", p.Flags[0].Short)
	assert.Equal(t, fields.TypeString, p.Flags[0].Type)
	assert.Equal(t, "foobar", p.Flags[0].Value)
}

func TestSingleFlagDefaultValue(t *testing.T) {
	pdTest := fields.New("test",
		fields.TypeString,
		fields.WithDefault("foobar"),
		fields.WithHelp("testing help"),
	)
	d := cmds.NewCommandDescription("test",
		cmds.WithFlags(
			pdTest,
		),
	)
	p := NewProgramFromCapture(d, makeParsedDefaultLayer(d, fields.NewFieldValues(fields.WithFieldValue(pdTest, "test", "foobar"))))

	assert.Equal(t, "test", p.Name)
	assert.Equal(t, "", p.Description)
	assert.Len(t, p.Flags, 0)

	p = NewProgramFromCapture(d, makeParsedDefaultLayer(d, fields.NewFieldValues(fields.WithFieldValue(pdTest, "test", "foobar2"))))

	assert.Equal(t, "test", p.Name)
	assert.Equal(t, "", p.Description)
	assert.Len(t, p.Flags, 1)
	assert.Equal(t, "test", p.Flags[0].Name)
	assert.Equal(t, "testing help", p.Flags[0].Short)
	assert.Equal(t, fields.TypeString, p.Flags[0].Type)
	assert.Equal(t, "foobar2", p.Flags[0].Value)
}

func TestTwoFlags(t *testing.T) {
	pd1 := fields.New("test", fields.TypeString)
	pd2 := fields.New("test2", fields.TypeString)
	d := cmds.NewCommandDescription("test",
		cmds.WithFlags(
			pd1,
			pd2,
		),
	)

	p := NewProgramFromCapture(
		d,
		makeParsedDefaultLayer(d, fields.NewFieldValues(
			fields.WithFieldValue(pd1, "test", "foobar"),
			fields.WithFieldValue(pd2, "test2", "foobar2"),
		)))

	assert.Equal(t, "test", p.Name)
	assert.Equal(t, "", p.Description)
	assert.Len(t, p.Flags, 2)
	assert.Equal(t, "test", p.Flags[0].Name)
	assert.Equal(t, "", p.Flags[0].Short)
	assert.Equal(t, fields.TypeString, p.Flags[0].Type)
	assert.Equal(t, "foobar", p.Flags[0].Value)
	assert.Equal(t, "test2", p.Flags[1].Name)
	assert.Equal(t, "", p.Flags[1].Short)
	assert.Equal(t, fields.TypeString, p.Flags[1].Type)
	assert.Equal(t, "foobar2", p.Flags[1].Value)
}

func TestSingleArg(t *testing.T) {
	pd := fields.New("test", fields.TypeString)
	d := cmds.NewCommandDescription("test",
		cmds.WithArguments(
			pd,
		),
	)
	p := NewProgramFromCapture(
		d,
		makeParsedDefaultLayer(d,
			fields.NewFieldValues(
				fields.WithFieldValue(pd, "test", "foobar"))))

	assert.Equal(t, "test", p.Name)
	assert.Equal(t, "", p.Description)
	assert.Len(t, p.Args, 1)
	assert.Equal(t, "test", p.Args[0].Name)
	assert.Equal(t, "", p.Args[0].Short)
	assert.Equal(t, fields.TypeString, p.Args[0].Type)
	assert.Equal(t, "foobar", p.Args[0].Value)
}

func TestTwoArgsTwoFlags(t *testing.T) {
	pd1 := fields.New("test", fields.TypeString)
	pd2 := fields.New("test2", fields.TypeString)
	pd3 := fields.New("test3", fields.TypeString)
	pd4 := fields.New("test4", fields.TypeString)
	d := cmds.NewCommandDescription("test",
		cmds.WithArguments(
			pd1,
			pd2,
		),
		cmds.WithFlags(
			pd3,
			pd4,
		),
	)
	p := NewProgramFromCapture(
		d,
		makeParsedDefaultLayer(d, fields.NewFieldValues(
			fields.WithFieldValue(pd1, "test", "foobar"),
			fields.WithFieldValue(pd2, "test2", "foobar2"),
			fields.WithFieldValue(pd3, "test3", "foobar3"),
			fields.WithFieldValue(pd4, "test4", "foobar4"),
		)),
	)

	assert.Equal(t, "test", p.Name)
	assert.Equal(t, "", p.Description)
	assert.Len(t, p.Args, 2)
	assert.Equal(t, "test", p.Args[0].Name)
	assert.Equal(t, "", p.Args[0].Short)
	assert.Equal(t, fields.TypeString, p.Args[0].Type)
	assert.Equal(t, "foobar", p.Args[0].Value)
	assert.Equal(t, "test2", p.Args[1].Name)
	assert.Equal(t, "", p.Args[1].Short)
	assert.Equal(t, fields.TypeString, p.Args[1].Type)
	assert.Equal(t, "foobar2", p.Args[1].Value)
	assert.Len(t, p.Flags, 2)
	assert.Equal(t, "test3", p.Flags[0].Name)
	assert.Equal(t, "", p.Flags[0].Short)
	assert.Equal(t, fields.TypeString, p.Flags[0].Type)
	assert.Equal(t, "foobar3", p.Flags[0].Value)
	assert.Equal(t, "test4", p.Flags[1].Name)
	assert.Equal(t, "", p.Flags[1].Short)
	assert.Equal(t, fields.TypeString, p.Flags[1].Type)
	assert.Equal(t, "foobar4", p.Flags[1].Value)
}

func TestSingleLayer(t *testing.T) {
	pd := fields.New("test", fields.TypeString)
	layer, err2 := schema.NewSection("test-layer", "test-layer",
		schema.WithFields(
			pd,
		),
	)
	require.NoError(t, err2)

	d := cmds.NewCommandDescription("test",
		cmds.WithLayersList(
			layer,
		),
	)

	ret := values.New()
	ret.Set("test-layer", &values.SectionValues{
		Section: layer,
		Fields: fields.NewFieldValues(
			fields.WithFieldValue(pd, "test", "foobar"))})

	p := NewProgramFromCapture(d, ret)

	assert.Equal(t, "test", p.Name)
	assert.Equal(t, "", p.Description)
	assert.Len(t, p.Flags, 1)
	assert.Equal(t, "test", p.Flags[0].Name)
	assert.Equal(t, "", p.Flags[0].Short)
}
