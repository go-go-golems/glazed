---
Title: Outputting multiple output files
Slug: output-multiple-files
Command: glaze
Short: |
  ```
  glaze csv misc/test-data/employees.csv --output csv \
     --output-file /tmp/employee.csv --output-multiple-files
  ```
Topics:
- templates
- output
Commands:
- json
- yaml
- csv
Flags:
- output-multiple-files
- output-file-template
- output-file
IsTemplate: false
IsTopLevel: false
ShowPerDefault: false
SectionType: Example
---

You can output each row into a separate file using the `--output-multiple-files` flag.

When using this flag, you must also specify the `--output-file` flag or the `--output-file-template` flag.

### Using the `--output-file` flag

When using the `--output-file` flag, the output file name will be the basename with the row index appended.
For example, if you specify `--output-file /tmp/employee.csv`, 
the output files will be named `/tmp/employee-0.csv`, `/tmp/employee-1.csv`, etc.


``` 
❯ glaze csv misc/test-data/employees.csv --output csv --output-file /tmp/employee.csv --output-multiple-files             
Written output to /tmp/employee-0.csv
Written output to /tmp/employee-1.csv
Written output to /tmp/employee-2.csv
Written output to /tmp/employee-3.csv
Written output to /tmp/employee-4.csv
Written output to /tmp/employee-5.csv
Written output to /tmp/employee-6.csv
Written output to /tmp/employee-7.csv
Written output to /tmp/employee-8.csv
Written output to /tmp/employee-9.csv

❯ cat /tmp/employee-0.csv 
First Name,Last Name,Title,Salary
John,Doe,Manager,"$75,000"
```


### Using the `--output-file-template` flag

When using the `--output-file-template` flag, the output file name as a template. 
The row values are available as variables in the template, as well as the row index as the variable `rowIndex`.

``` 
❯ glaze csv misc/test-data/employees.csv --output csv --output-file /tmp/employee.csv --output-multiple-files --output-file-template '/tmp/employee-{{.Title}}-{{.rowIndex}}.csv' 
Written output to /tmp/employee-Manager-0.csv
Written output to /tmp/employee-Software Engineer-1.csv
Written output to /tmp/employee-Sales Representative-2.csv
Written output to /tmp/employee-Marketing Manager-3.csv
Written output to /tmp/employee-Web Developer-4.csv
Written output to /tmp/employee-Accountant-5.csv
Written output to /tmp/employee-HR Manager-6.csv
Written output to /tmp/employee-Software Developer-7.csv
Written output to /tmp/employee-Customer Service Representative-8.csv
Written output to /tmp/employee-Data Analyst-9.csv

❯ cat /tmp/employee-Software\ Developer-7.csv 
First Name,Last Name,Title,Salary
Sarah,Taylor,Software Developer,"$80,000"

```