Below is a reference for how to create YAML command files that define a **SqlCommand** (or similar command objects) without discussing the actual query text or any layout/alias configuration. These YAML files describe:

- Command **metadata** (like environment or a short description)
- A set of **flags** (and optionally arguments) that the command will accept
- Basic usage instructions and help text

---

## Overall YAML Structure

A typical YAML command file contains the following top-level keys:

1. **name**:  
   The internal or CLI name of the command.  
   *Example*: `name: top-customers`

2. **metadata**:  
   A dictionary containing extra metadata about the command.  
   Commonly, this might include:
   - `environment: ttc_analytics` or `environment: ttc_prod`
   - `note:` or other fields relevant to the environment or context
   
   *Example*:
   ```yaml
   metadata:
     environment: ttc_analytics
     note: This command is for analytics data only
   ```

3. **short**:  
   A short description (one line) of what the command does.  
   *Example*: `short: Report on container/gallon AB test results.`

4. **long** (optional):  
   A longer multi-line description.

5. **flags** and/or **arguments**:  
   Parameters can be defined either as flags (with -- prefix) or positional arguments. Both use the same parameter definition structure, just with different usage patterns. Each entry describes one parameter with fields such as:
   - `name` (required): The parameter name.
   - `type` (required): The parameter type.
   - `help` (optional): A short help string describing usage.
   - `default` (optional): The default value if none is provided.
   - `required` (optional, boolean): Indicates if this parameter must be supplied.
   - `choices` (optional, array): If the type is a choice or choiceList, valid strings must be one of these choices.

   **Note**:   
   - If `type` is `bool`, then including `required: false` simply means it defaults to `false` unless passed in some way.  
   - If you set `type: date`, you can often provide `default: 2024-01-01` (or any valid date format).  
   - For a `type: choice` or `type: choiceList`, you provide a `choices: [ ... ]` list.  
   - For lists, you might see `type: stringList` or `type: intList` or `type: floatList`.  
   - For single numeric parameters, `type` can be `int` or `float`.  
   - For textual parameters, `type` can be `string`.  
   - `help` is free-form text but should be short.
   - Arguments are positional parameters that don't use -- prefix
   - Both flags and arguments support the same types and configuration options
   - Arguments are typically required unless they have a default value

**Example** (with no query portion shown):

```yaml
name: container-gallon-ab-test
metadata:
  environment: ttc_analytics
short: Container/gallon wording a/b test result.
flags:
  - name: from
    type: date
    help: Start date (inclusive)
    default: 2024-04-30
  - name: to
    type: date
    help: End date (inclusive)
  - name: order_by
    type: string
    default: "cohort"
    help: Columns to order results by
```

Here:
- **name** is `container-gallon-ab-test`.
- **metadata** has just `environment: ttc_analytics`.
- **short** is a single-line summary of what the command is about.
- **flags** includes three parameters:
  1. A date flag `from`, with default `2024-04-30`.
  2. A date flag `to`, no default, not required → optional.
  3. A string flag `order_by`, defaulting to `"cohort"`.

---

## Flag Definition Fields

Each item under `flags:` or `arguments:` is typically structured like:

```yaml
- name: <parameter-name>
  type: <parameter-type>
  help: <short description for usage>
  default: <default value>
  required: <true|false>
  choices: [ ... ]
```

Where:

- **name**  
  Unique parameter name (e.g. `limit`, `from`, `status`), used at the command line as `--limit=...` or similar.

- **type**  
  Indicates how this parameter is interpreted. Common types:

  1. **string**: A single string (e.g., `--status=wc-completed`).
  2. **int**: An integer (e.g., `--limit=10`).
  3. **float**: A floating-point number (e.g., `--min_aov=15.75`).
  4. **bool**: A boolean (e.g., `--used` or `--used=false`).
  5. **date**: Typically expects a date string or natural date expression (`2024-01-01`).
  6. **stringList**: A list of strings (e.g., `--status=wc-completed --status=wc-processing` or `--status=wc-completed,wc-processing`).
  7. **intList**: A list of integers.
  8. **floatList**: A list of floats.
  9. **choice**: A string that must match one of a given set of `choices:`.
  10. **choiceList**: A list of strings, each of which must match one of the `choices:`.
  11. **keyValue**: Key-value pairs (e.g., `--header='Content-Type:application/json'`).
  12. **file**: A single file input, providing file data and metadata.
  13. **fileList**: A list of file inputs.
  14. (Less common) **stringFromFile**, **objectFromFile**, **stringListFromFile**, **objectListFromFile**: Indicate the parameter is read from a file or multiple files. Typically used in advanced scenarios.

- **help**  
  A short description that explains what this flag does.

- **default** (optional)  
  A default value if the user does not provide one.  
  *Example*: `default: false` for a boolean, `default: 100` for an integer, or `default: 2024-01-01` for a date.

- **required** (optional)  
  Set to `true` if it must be provided. By default, it is `false`. If missing in user input and no default is given, the command parser may fail.

- **choices** (optional)  
  Only used when `type: choice` or `type: choiceList`, listing valid string options.  
  *Example*:
  ```yaml
  - name: group_by
    type: choice
    help: Grouping level
    default: year
    choices: [year, month, all-time]
  ```

### Examples of Flag Definitions

1. **Simple string with a default:**
   ```yaml
   - name: order_by
     type: string
     default: "cohort"
     help: Sort results by a column
   ```

2. **Integer with a default and not required:**
   ```yaml
   - name: limit
     type: int
     default: 100
     help: Maximum number of rows
   ```

3. **Boolean flag with no default (thus false if omitted):**
   ```yaml
   - name: with_orders
     type: bool
     help: Select coupons that have orders
     default: false
   ```

4. **Date range flags:**
   ```yaml
   - name: from
     type: date
     help: Start date
     default: 2024-01-01

   - name: to
     type: date
     help: End date
   ```

5. **List of strings:**
   ```yaml
   - name: status
     type: stringList
     default: ['wc-scheduled', 'wc-completed', 'wc-processing']
     help: Order statuses to include
   ```

6. **Choice**:
   ```yaml
   - name: group_by
     type: choice
     choices: [year, month, all-time]
     default: year
     help: Group results by year, month, or all-time
   ```

7. **Choice list**:
   ```yaml
   - name: years_with_orders
     type: choiceList
     help: Limit results to these specific years
     choices: ["2020","2021","2022","2023"]
   ```

8. **Integer list**:
   ```yaml
   - name: order_id
     type: intList
     help: List of order IDs
   ```

9. **File input**:
   ```yaml
   - name: config
     type: file
     help: Configuration file to process
     required: true
   ```

10. **List of files**:
    ```yaml
   - name: input_files
     type: fileList
     help: List of input files to process
    ```

## Example with Both Flags and Arguments

Below is an example showing both flags and positional arguments:

```yaml
name: process-files
metadata:
  environment: ttc_analytics
short: Process input files with given configuration.
arguments:
  - name: input_file
    type: file
    help: Primary input file to process
    required: true
  
  - name: output_path
    type: string
    help: Where to write the results
    required: true

flags:
  - name: config
    type: file
    help: Optional configuration file
    
  - name: verbose
    type: bool
    help: Enable verbose output
    default: false

  - name: format
    type: choice
    choices: [json, yaml, csv]
    default: json
    help: Output format to use
```

In this example:
- Two required positional arguments: input_file and output_path
- Three optional flags: --config, --verbose, and --format
- Both flags and arguments use the same parameter definition structure
- The file type is used for file inputs, providing access to file data and metadata

---

## Putting It All Together: A Minimal Example

Below is a concise YAML file that demonstrates a command definition with flags:

```yaml
name: user-stats
metadata:
  environment: ttc_analytics
short: Compute basic user statistics based on date range.
flags:
  - name: from
    type: date
    help: Start date (inclusive)
    default: 2023-01-01

  - name: to
    type: date
    help: End date (inclusive)
    # no default => optional

  - name: include_inactive
    type: bool
    help: Whether to include inactive users
    default: false

  - name: limit
    type: int
    help: Maximum number of users returned
    default: 50

  - name: group_by
    type: choice
    choices: [month, year]
    default: month
    help: Granularity of the statistics
```

- The essential fields are `name`, `metadata`, `short`, and the `flags` list (with each flag's name, type, help, etc.).

---

## Summary of Key Points

1. **name** (string) – Uniquely identifies the command.  
2. **metadata** (object) – Additional info like `environment: ttc_analytics`.  
3. **short** (string) – Brief description.  
4. **flags** (list of objects) – Each entry has:
   - `name`: parameter/flag name
   - `type`: see typical parameter types (e.g., `string`, `int`, `bool`, `date`, etc.)
   - `help`: short explanatory text
   - `default`: (optional) default value
   - `required`: (optional) boolean
   - `choices`: (optional) for `choice`/`choiceList`  
   
Use this structure whenever you define a new YAML command file. The system that loads it will parse these fields and make your parameters available under the indicated names.

That's all that's needed for the command *description* itself.

## How This YAML Relates to a CommandDescription

When loading this YAML, the system essentially constructs a `CommandDescription` (in code) that captures:

1. **Command Identity**  
   - `name: coupons` → The `CommandDescription.Name` field.  
   - `short: List coupons...` → The `CommandDescription.Short`.  
   - `long: > ...` (optional) → The `CommandDescription.Long`.  
   - `metadata:` can become part of `CommandDescription.AdditionalData` or environment context.

2. **Flags (Parameters)**  
   - Under `flags:`, each item (e.g. `- name: name`, `type: stringList`, etc.) is turned into a `ParameterDefinition`.
   - In Go code, this corresponds to something like:

     ```go
     parameters.NewParameterDefinition(
       "name",
       parameters.ParameterTypeStringList,
       parameters.WithHelp("List of coupon names"),
     )
     ```

   - The system aggregates these into a "default layer" of parameters.  
   - `required: false`, `default: ...`, and `help: ...` become `WithRequired(false)`, `WithDefault(...)`, `WithHelp(...)` in the `ParameterDefinition` construction.

Hence, the **top-level YAML keys**—`name`, `metadata`, `short`, `long`, and `flags`—**map to** the **fields and parameter layers** in a `CommandDescription` object. The `query:` would likewise be used by a specialized command that actually runs a SQL query, but the fundamental principle of turning flags into typed parameters remains the same.