# Field Types Example

This example demonstrates all field types available in the glazed framework.

## Usage

```bash
# Build the example
go build -o field-types .

# Show help to see all available fields
./field-types --help

# Run with default values
./field-types

# Test basic types
./field-types \
  --string-field "hello world" \
  --secret-field "my-secret" \
  --integer-field 100 \
  --float-field 2.71828 \
  --bool-field=false \
  --date-field "2024-12-25" \
  --choice-field option2

# Test list types
./field-types \
  --string-list-field item1,item2,item3 \
  --integer-list-field 10,20,30 \
  --float-list-field 1.1,2.2,3.3 \
  --choice-list-field red,green

# Test file types
./field-types \
  --file-field sample.json \
  --file-list-field sample.json,sample.yaml \
  --string-from-file-field sample-text.txt \
  --string-list-from-file-field sample-lines.txt \
  --object-from-file-field sample.json \
  --object-list-from-file-field sample-list.json

# Test key-value type
./field-types \
  --key-value-field key1:value1,key2:value2

# Load key-value from file
./field-types \
  --key-value-field @config.yaml
```

## Field Types Demonstrated

### Basic Types
- **string**: Simple text values
- **secret**: Text values that are masked when displayed (***) 
- **integer**: Whole numbers
- **float**: Decimal numbers
- **bool**: Boolean true/false values
- **date**: Date/time values (RFC3339 or natural language)
- **choice**: Single value from predefined options

### List Types  
- **string-list**: Multiple text values
- **integer-list**: Multiple integers
- **float-list**: Multiple decimal numbers
- **choice-list**: Multiple values from predefined options

### File Types
- **file**: Load file with metadata (path, size, type, etc.)
- **file-list**: Load multiple files with metadata
- **string-from-file**: Load file content as a single string
- **string-from-files**: Load and concatenate multiple files as string
- **string-list-from-file**: Load file lines as string list
- **string-list-from-files**: Load lines from multiple files as string list
- **object-from-file**: Parse JSON/YAML file as object
- **object-list-from-file**: Parse JSON/YAML file as object list  
- **object-list-from-files**: Parse multiple files and merge object lists

### Key-Value Type
- **key-value**: Map of key-value pairs (key:value format or @file)

## Sample Files

- `sample.json` - JSON object for testing object fields
- `sample.yaml` - YAML data for testing YAML parsing
- `sample-list.json` - JSON array for testing list fields
- `sample-text.txt` - Multi-line text for string fields
- `sample-lines.txt` - Line-by-line text for list fields
- `config.yaml` - Configuration file for key-value fields

## Output

The program displays all parsed field values in a structured format, demonstrating how each type is processed and what the final values look like.

Note that secret fields will show as `***` in the output to protect sensitive data.
