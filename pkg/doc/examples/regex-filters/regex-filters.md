---
Title: Filter fields using Regular Expressions
Slug: regex-filters
Short: |
  ```
  glaze json misc/test-data/sample.json --input-is-array --regex-fields "^name_.*" --regex-filters ".*_last$"
  ```
Topics:
  - regex-filters
  - regex-fields
Commands:
  - yaml
  - json
  - csv
Flags:
  - regex-fields
  - regex-filters
IsTemplate: false
IsTopLevel: true
ShowPerDefault: false
SectionType: Example
---

You can use regular expressions to include or exclude fields from the output. This is useful when dealing with a large number of fields or fields with dynamic names.

The `--regex-fields` flag allows you to specify regex patterns for fields you want to **include**.
The `--regex-filters` flag allows you to specify regex patterns for fields you want to **exclude**.

These flags can be combined with the standard `--fields` and `--filter` flags. The precedence rules are complex but generally, filters (both standard and regex) take precedence over includes. Exact matches (`--fields`, `--filter`) usually have higher priority than regex or prefix matches.

Let's use the following sample data (`misc/test-data/sample.json`):

```json
[
  {
    "id": 1,
    "name_first": "John",
    "name_last": "Doe",
    "address_street": "123 Main St",
    "address_city": "Anytown",
    "email": "john.doe@example.com",
    "phone_home": "555-1234",
    "phone_work": "555-5678"
  },
  {
    "id": 2,
    "name_first": "Jane",
    "name_last": "Smith",
    "address_street": "456 Oak Ave",
    "address_city": "Otherville",
    "email": "jane.smith@example.com",
    "phone_home": "555-4321",
    "phone_work": "555-8765"
  }
]
```

---

## Examples

### 1. Include fields starting with "name\_"

```
❯ glaze json misc/test-data/sample.json --input-is-array --regex-fields "^name_.*"
+------------+-----------+
| name_first | name_last |
+------------+-----------+
| John       | Doe       |
| Jane       | Smith     |
+------------+-----------+
```

### 2. Include fields containing "address"

```
❯ glaze json misc/test-data/sample.json --input-is-array --regex-fields "address"
+----------------+--------------+
| address_street | address_city |
+----------------+--------------+
| 123 Main St    | Anytown      |
| 456 Oak Ave    | Otherville   |
+----------------+--------------+
```

### 3. Exclude fields ending with "\_work"

```
❯ glaze json misc/test-data/sample.json --input-is-array --regex-filters "_work$"
+----+------------+-----------+----------------+--------------+------------------------+------------+
| id | name_first | name_last | address_street | address_city | email                  | phone_home |
+----+------------+-----------+----------------+--------------+------------------------+------------+
|  1 | John       | Doe       | 123 Main St    | Anytown      | john.doe@example.com   | 555-1234   |
|  2 | Jane       | Smith     | 456 Oak Ave    | Otherville   | jane.smith@example.com | 555-4321   |
+----+------------+-----------+----------------+--------------+------------------------+------------+

```

### 4. Exclude fields related to "phone"

```
❯ glaze json misc/test-data/sample.json --input-is-array --regex-filters "^phone_.*"
+----+------------+-----------+----------------+--------------+------------------------+
| id | name_first | name_last | address_street | address_city | email                  |
+----+------------+-----------+----------------+--------------+------------------------+
|  1 | John       | Doe       | 123 Main St    | Anytown      | john.doe@example.com   |
|  2 | Jane       | Smith     | 456 Oak Ave    | Otherville   | jane.smith@example.com |
+----+------------+-----------+----------------+--------------+------------------------+
```

### 5. Combine Regex Fields and Filters: Include "name\_" fields, exclude "\_last"

```
❯ glaze json misc/test-data/sample.json --input-is-array --regex-fields "^name_.*" --regex-filters "_last$"
+------------+
| name_first |
+------------+
| John       |
| Jane       |
+------------+
```

### 6. Combine Regex and Standard Flags: Include "address" fields via regex, but specifically filter out "address_city"

```
❯ glaze json misc/test-data/sample.json --input-is-array --regex-fields "address" --filter address_city
+----------------+
| address_street |
+----------------+
| 123 Main St    |
| 456 Oak Ave    |
+----------------+
```

### 7. Combine Regex and Standard Flags: Include "id" specifically, and all "phone" fields via regex

```
❯ glaze json misc/test-data/sample.json --input-is-array --fields id --regex-fields "^phone_.*"
+----+------------+------------+
| id | phone_home | phone_work |
+----+------------+------------+
|  1 | 555-1234   | 555-5678   |
|  2 | 555-4321   | 555-8765   |
+----+------------+------------+
```
