# Parameter Types Example

This example demonstrates all parameter types available in the glazed framework.

## Usage

```bash
# Build the example
go build -o parameter-types .

# Show help to see all available parameters
./parameter-types --help

# Run with default values
./parameter-types

# Test basic types
./parameter-types \
  --string-param "hello world" \
  --secret-param "my-secret" \
  --integer-param 100 \
  --float-param 2.71828 \
  --bool-param=false \
  --date-param "2024-12-25" \
  --choice-param option2

# Test list types
./parameter-types \
  --string-list-param item1,item2,item3 \
  --integer-list-param 10,20,30 \
  --float-list-param 1.1,2.2,3.3 \
  --choice-list-param red,green

# Test file types
./parameter-types \
  --file-param sample.json \
  --file-list-param sample.json,sample.yaml \
  --string-from-file-param sample-text.txt \
  --string-list-from-file-param sample-lines.txt \
  --object-from-file-param sample.json \
  --object-list-from-file-param sample-list.json

# Test key-value type
./parameter-types \
  --key-value-param key1:value1,key2:value2

# Load key-value from file
./parameter-types \
  --key-value-param @config.yaml
```

## Parameter Types Demonstrated

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

- `sample.json` - JSON object for testing object parameters
- `sample.yaml` - YAML data for testing YAML parsing
- `sample-list.json` - JSON array for testing list parameters
- `sample-text.txt` - Multi-line text for string parameters
- `sample-lines.txt` - Line-by-line text for list parameters
- `config.yaml` - Configuration file for key-value parameters

## Output

The program displays all parsed parameter values in a structured format, demonstrating how each type is processed and what the final values look like.

Note that secret parameters will show as `***` in the output to protect sensitive data.
