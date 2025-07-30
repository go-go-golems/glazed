# Help System Example

This example demonstrates the Glazed help system functionality based on the documentation in `/pkg/doc/topics/01-help-system.md`.

## What This Example Tests

1. **Help System Creation**: Using `help.NewHelpSystem()` 
2. **Loading Documentation**: From embedded filesystem using `LoadSectionsFromFS()`
3. **Programmatic Section Addition**: Adding sections directly to the help system
4. **DSL Queries**: Testing various query patterns including:
   - Type-based queries (`type:example`)
   - Boolean logic (`type:example AND topic:testing`)
   - Text search (`"markdown"`)
5. **Individual Section Retrieval**: Using `GetSectionWithSlug()`
6. **Cobra Integration**: Example of how to integrate with Cobra commands

## Sample Documentation

The example includes sample documentation in the `docs/` directory:

- `database-tutorial.md` - Tutorial section type
- `json-example.md` - Example section type  
- `configuration-topic.md` - GeneralTopic section type
- `data-pipeline-application.md` - Application section type

Each file demonstrates the YAML frontmatter structure used by the help system.

## Running the Example

```bash
go run main.go
```

## Key Features Demonstrated

- **Section Types**: All four types (GeneralTopic, Example, Application, Tutorial)
- **Metadata**: Topics, Commands, Flags for filtering
- **Display Options**: IsTopLevel, ShowPerDefault, Order
- **Query DSL**: Boolean logic and field-based queries
- **Cobra Integration**: Basic help command setup

## Documentation Accuracy

This example verifies that the API described in `/pkg/doc/topics/01-help-system.md` is accurate and functional.
