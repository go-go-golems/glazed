# RFC - glaze CLI tool

## Changelog

### 2022-11-19 - manuel

- Created document
- Brainstorm

## Overview

The `glaze` tool is both meant to:
- showcase the glazed library as used by normal users
- showcase the glazed library when used by developers
- be a useful general purpose data transformation / visualization tool

`glaze` is able to read in files in the following format:
- CSV
- JSON
- SQLite
- YAML

Every middleware provided in the `glazed` library will be used,
either by having a specific command dedicated to it (when needed with synthetic data)
or by being used for some of the more user oriented transformation commands.

## User side

The `glaze` application should allow a user to:
- read in data from files
  - json - #13 [x]
  - csv - #14 [x]
  - sqlite - #15 [x]
  - yaml - #16 [x]
- write out data
  - json [x]
  - csv [x]
  - sqlite
  - yaml [x]
- select and order columns

## Developer side

The `glaze` sourcecode should showcase every interesting feature of `glazed`.

### Initialization 

- Using the flag helpers
  - This might require a separate command line tool per backend altogether
- Loading configurations to file
- Saving configurations to file

### Output formatters

- Output as CSV [x]
- Output as human readable table [x]
- Output as json [x]
- Output as SQLite
- Output as yaml [x]

### Middlewares

#### Table middlewares

- Fields filter [x]
- Flatten Object Paths [x]
- Preserve Column Order
- Reorder Column
- Sort Columns

#### Object middlewares

GO template middlewares
- Row Go Template
- Object Go Template
- File Go Template
