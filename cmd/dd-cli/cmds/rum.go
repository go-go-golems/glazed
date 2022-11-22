package cmds

import (
	"encoding/json"
	"fmt"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/araddon/dateparse"
	_ "github.com/araddon/dateparse"
	_ "github.com/scylladb/termtables"
	"github.com/spf13/cobra"
	"glazed/pkg/cli"
	"glazed/pkg/formatters"
	"glazed/pkg/middlewares"
	"glazed/pkg/types"
	"os"
	"strings"
	"unicode/utf8"
)

var RumCmd = cobra.Command{
	Use:   "rum",
	Short: "Query DataDog RUM",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

type Action struct {
	Name       string
	Attributes map[string]interface{}
	Context    interface{}
}

var listActionsCmd = cobra.Command{
	Use:   "ls-actions",
	Short: "List RUM actions",
	Long: `List RUM actions.

This command allows you to specify action names to look for, using comma-separated lists:

    dd-cli rum ls-actions --action "action1,action2"

The output can be JSON, in which case you get the entire action objects:

	dd-cli rum ls-actions --action action1 --count 23 --output json

When using the output types other than JSON, the data in the "context" field gets turned into 
individual columns. The other objects are ignored.

You can also specify fields to include in the output, using comma-separated lists:

	dd-cli rum ls-actions --fields "name,page.url"

You can also specify fields to remove from the output, using comma-separated lists:

	dd-cli rum ls-actions --filter "page.url"

Here is a more complex example:
	dd-cli rum ls-actions \
 		--action filters-search \
        --from 2022/11/10 
		--fields name,query.q,query.,pagination.total_hits \
		--filter query.category \
		--table-format markdown \
		--count 50 


With CSV output:
	dd-cli rum ls-actions \
				--action filters-search --from 2022/11/10 \
               	--count 50 \
				--table-format csv --csv-separator '|' --with-headers=false

With a single line template:
	dd-cli rum ls-actions \
				--action filters-search --from 2022/11/10 \
				--count 12 \
				--template '{{.name}} - {{.pagination.total_hits}}'

With field templates:
	dd-cli rum ls-actions \
				--action filters-search --from 2022/11/10 --count 12 \
				--table-format markdown \
				--template-field 'yo:{{.name | ToUpper }}' \
				--template-field 'hits:{{.pagination.total_hits}}' \
				--fields 'name,yo,hits'

Give a yaml file:

	yo: "{{ .name | ToUpper }}"
	hits: "{{ .pagination.total_hits }}"

	dd-cli rum ls-actions \
				--action filters-search --from 2022/11/10 --count 12 \
				--table-format markdown \
				--template-field '@misc/rum-action-field-templates.yaml' \
				--template-field 'double:{{.name}}-{{.name}}' \
				--fields 'name,double,yo,hits'

	`,
	Run: func(cmd *cobra.Command, args []string) {
		from := cmd.Flag("from").Value.String()
		to := cmd.Flag("to").Value.String()
		action := cmd.Flag("action").Value.String()
		actionNames := []string{}
		if action != "" {
			actionNames = strings.Split(action, ",")
		}

		filter := datadogV2.NewRUMQueryFilter()
		query := "@type:action"
		if action != "" {
			query += fmt.Sprintf(" @action.name:(%s)", strings.Join(actionNames, " OR "))
		}

		filter.SetQuery(query)
		if from != "" {
			t, err := dateparse.ParseLocal(from)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Could not parse %s: %s", from, err)
				return
			}
			filter.SetFrom(fmt.Sprintf("%d", t.Unix()))
		}
		if to != "" {
			t, err := dateparse.ParseLocal(to)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Could not parse %s: %s", t, err)
				return
			}
			filter.SetTo(fmt.Sprintf("%d", t.Unix()))
		}

		body := datadogV2.NewRUMSearchEventsRequest()
		body.SetFilter(*filter)
		configuration := datadog.NewConfiguration()
		apiClient := datadog.NewAPIClient(configuration)
		rumApi := datadogV2.NewRUMApi(apiClient)

		events, cancel, err := rumApi.SearchRUMEventsWithPagination(CmdContext, *body)
		if err != nil {
			panic(err)
		}
		defer cancel()

		actions := []Action{}

		count, err := cmd.Flags().GetInt("count")
		if err != nil {
			panic(err)
		}

		for i := 0; i < count; i++ {
			e := <-events
			attrs := e.GetAttributes().Attributes
			action, hasAction := attrs["action"].(map[string]interface{})
			if !hasAction {
				continue
			}
			actions = append(actions, Action{
				Name:       action["name"].(string),
				Attributes: attrs,
				Context:    attrs["context"],
			})
		}

		// these are the flags used for the table output
		output := cmd.Flag("output").Value.String()
		_ = cmd.Flag("output-file").Value.String()

		fieldStr := cmd.Flag("fields").Value.String()
		filters := []string{}
		fields := []string{}
		if fieldStr != "" {
			fields = strings.Split(fieldStr, ",")
		}
		filterStr := cmd.Flag("filter").Value.String()
		if filterStr != "" {
			filters = strings.Split(filterStr, ",")
		}

		tableFormat := cmd.Flag("table-format").Value.String()

		// templates get applied before flattening
		var templates map[types.FieldName]string

		templateArgument, _ := cmd.Flags().GetString("template")
		if templateArgument != "" {
			templates = map[types.FieldName]string{}
			templates["_0"] = templateArgument
		} else {
			templateFields, _ := cmd.Flags().GetStringSlice("template-field")
			templates, err = cli.ParseTemplateFieldArguments(templateFields)
			if err != nil {
				panic(err)
			}
		}

		if templateArgument != "" {
			// if a template is specified, we only print out the individual template lines
			flattenedActions, err := flattenActions(actions, templates)
			if err != nil {
				panic(err)
			}
			for _, action := range flattenedActions {
				fmt.Println(action["_0"])
			}
		} else {
			if output == "json" {
				jsonBytes, err := json.MarshalIndent(actions, "", "  ")
				if err != nil {
					panic(err)
				}
				fmt.Println(string(jsonBytes))
			} else if output == "table" {
				withHeaders, _ := cmd.Flags().GetBool("with-headers")
				var of formatters.OutputFormatter
				if tableFormat == "csv" {
					csvSeparator, _ := cmd.Flags().GetString("csv-separator")

					csvOf := formatters.NewCSVOutputFormatter()
					csvOf.WithHeaders = withHeaders
					r, _ := utf8.DecodeRuneInString(csvSeparator)
					csvOf.Separator = r
					of = csvOf
				} else if tableFormat == "tsv" {
					tsvOf := formatters.NewTSVOutputFormatter()
					tsvOf.WithHeaders = withHeaders
					of = tsvOf
				} else {
					of = formatters.NewTableOutputFormatter(tableFormat)
				}

				of.AddTableMiddleware(middlewares.NewFlattenObjectMiddleware())
				of.AddTableMiddleware(middlewares.NewFieldsFilterMiddleware(fields, filters))
				of.AddTableMiddleware(middlewares.NewSortColumnsMiddleware())
				if len(fields) == 0 {
					of.AddTableMiddleware(middlewares.NewReorderColumnOrderMiddleware([]types.FieldName{"name"}))
				} else {
					of.AddTableMiddleware(middlewares.NewReorderColumnOrderMiddleware(fields))
				}

				flattenedActions, err := flattenActions(actions, templates)
				if err != nil {
					panic(err)
				}
				for _, action := range flattenedActions {
					of.AddRow(&types.SimpleRow{Hash: action})
				}

				s, err := of.Output()
				if err != nil {
					panic(err)
				}

				fmt.Println(s)
			} else if output == "sqlite" {

			}
		}

	},
}

func flattenActions(
	actions []Action,
	templates map[types.FieldName]string,
) ([]map[string]interface{}, error) {
	ret := []map[string]interface{}{}
	var gtmw *middlewares.ObjectGoTemplateMiddleware
	var err error
	if len(templates) > 0 {
		gtmw, err = middlewares.NewObjectGoTemplateMiddleware(templates)
		if err != nil {
			return nil, err
		}
	}

	for _, action := range actions {
		row := map[string]interface{}{}
		row["name"] = action.Name
		if action.Context != nil {
			context := action.Context.(map[string]interface{})
			if gtmw != nil {
				// we should pass action name to the context, or maybe actually the whole raw action
				context["name"] = action.Name
				res_, err := gtmw.Process(context)
				if err != nil {
					return nil, err
				}

				for k, v := range res_ {
					row[k] = v
				}
			}

			for k, v := range middlewares.FlattenMapIntoColumns(context) {
				row[k] = v
			}
		}
		ret = append(ret, row)
	}

	return ret, nil
}

func init() {
	RumCmd.AddCommand(&listActionsCmd)

	listActionsCmd.Flags().String("from", "", "From date (accepts variety of formats)")
	listActionsCmd.Flags().String("to", "", "To date (accepts variety of formats)")
	listActionsCmd.Flags().StringP("action", "a", "", "Action name")
	listActionsCmd.Flags().IntP("count", "c", 20, "Number of results to return")

	listActionsCmd.Flags().StringP("output", "o", "table", "Output format (table, json, sqlite)")
	listActionsCmd.Flags().StringP("output-file", "f", "", "Output file")
	listActionsCmd.Flags().String("table-format", "ascii", "Table format (ascii, markdown, html, csv, tsv)")
	listActionsCmd.Flags().Bool("with-headers", true, "Include headers in output (CSV, TSV)")
	listActionsCmd.Flags().String("csv-separator", ",", "CSV separator")

	listActionsCmd.Flags().String("template", "", "Go Template to use for single string")
	listActionsCmd.Flags().StringSlice("template-field", nil, "For table output, fieldName:template to create new fields, or @fileName to read field templates from a yaml dictionary")

	listActionsCmd.Flags().String("fields", "", "Fields to include in the output, default: all")
	listActionsCmd.Flags().String("filter", "", "Fields to remove from output")

}
