package cmds

import (
	"dd-cli/pkg"
	"encoding/json"
	"fmt"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/araddon/dateparse"
	_ "github.com/araddon/dateparse"
	_ "github.com/scylladb/termtables"
	"github.com/spf13/cobra"
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
	`,
	Run: func(cmd *cobra.Command, args []string) {
		from := cmd.Flag("from").Value.String()
		to := cmd.Flag("to").Value.String()
		action := cmd.Flag("action").Value.String()
		actionNames := []string{}
		if action != "" {
			actionNames = strings.Split(action, ",")
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

		if output == "json" {
			jsonBytes, err := json.MarshalIndent(actions, "", "  ")
			if err != nil {
				panic(err)
			}
			fmt.Println(string(jsonBytes))
		} else if output == "table" {
			withHeaders, _ := cmd.Flags().GetBool("with-headers")
			var of pkg.OutputFormatter
			if tableFormat == "csv" {
				csvSeparator, _ := cmd.Flags().GetString("csv-separator")

				csvOf := pkg.NewCSVOutputFormatter()
				csvOf.WithHeaders = withHeaders
				r, _ := utf8.DecodeRuneInString(csvSeparator)
				csvOf.Separator = r
				of = csvOf
			} else if tableFormat == "tsv" {
				tsvOf := pkg.NewTSVOutputFormatter()
				tsvOf.WithHeaders = withHeaders
				of = tsvOf
			} else {
				of = pkg.NewTableOutputFormatter(tableFormat)
			}

			of.AddMiddleware(pkg.NewFlattenObjectMiddleware())
			of.AddMiddleware(pkg.NewFieldsFilterMiddleware(fields, filters))
			of.AddMiddleware(pkg.NewSortColumnsMiddleware())
			if len(fields) == 0 {
				of.AddMiddleware(pkg.NewReorderColumnOrderMiddleware([]pkg.FieldName{"name"}))

			} else {
				of.AddMiddleware(pkg.NewReorderColumnOrderMiddleware(fields))
			}

			flattenedActions := flattenActions(actions)
			for _, action := range flattenedActions {
				of.AddRow(&pkg.SimpleRow{Hash: action})
			}

			s, err := of.Output()
			if err != nil {
				panic(err)
			}

			fmt.Println(s)
		} else if output == "sqlite" {

		}
	},
}

func flattenActions(actions []Action) []map[string]interface{} {
	ret := []map[string]interface{}{}

	for _, action := range actions {
		row := map[string]interface{}{}
		row["name"] = action.Name
		if action.Context != nil {
			context := action.Context.(map[string]interface{})
			for k, v := range pkg.FlattenMapIntoColumns(context) {
				row[k] = v
			}
		}
		ret = append(ret, row)
	}

	return ret
}

func init() {
	RumCmd.AddCommand(&listActionsCmd)

	listActionsCmd.Flags().String("from", "", "From date (accepts variety of formats)")
	listActionsCmd.Flags().String("to", "", "To date (accepts variety of formats)")

	listActionsCmd.Flags().StringP("output", "o", "table", "Output format (table, json, sqlite)")
	listActionsCmd.Flags().StringP("output-file", "f", "", "Output file")
	listActionsCmd.Flags().String("table-format", "ascii", "Table format (ascii, markdown, html, csv, tsv)")
	listActionsCmd.Flags().Bool("with-headers", true, "Include headers in output (CSV, TSV)")
	listActionsCmd.Flags().String("csv-separator", ",", "CSV separator")

	listActionsCmd.Flags().StringP("action", "a", "", "Action name")
	listActionsCmd.Flags().String("fields", "", "Fields to include in the output, default: all")
	listActionsCmd.Flags().String("filter", "", "Fields to remove from output")

	listActionsCmd.Flags().IntP("count", "c", 20, "Number of results to return")
}
