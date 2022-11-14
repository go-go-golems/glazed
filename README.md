# glazed - Output structured data in a variety of formats

> Add the icing to your structured data!

Glazed is a library that makes it easy to output structured data.
When programming, we have a rich understanding of the data we are working with,
yet when we output it to the user, we are forced to use a flat, unstructured format.

It tries to implement some of the ideas listed in
[14 great tips to make amazing command line applications](https://dev.to/wesen/14-great-tips-to-make-amazing-cli-applications-3gp3).

It is in early alpha, and will change. Contributions are welcome,
but this project is going to be an experimental playground for a while,
while I try to figure out what is possible and worth tackling.

## Features

With glazed, you can output object and table data in a rich variety of ways:

- export in CSV, JSON, markdown, html, text
- filter and rename columns
- use go templates to customize output
- output individual objects or rows as separate files

Glazed provides a variety of "middlewares" with which you can:

- flatten nested objects into rows
- create new fields based on go templates
- filter and reorder columns

For easy integration into your own tools, glazed provides:
- a simple API
- bindings for go command-line flag parsing
- cobra and viper libraries
- loading output configuration from a configuration file

Glazed also comes with the glaze tool which can be use for simple data manipulation
and rich terminal output, leveraging the glazed library.

## Getting started

### Using the glaze command line tool

First, [install the glaze tool](#Installation).

- Show 4-5 cool examples

### Developing with glazed

Write a tiny command line tool:
- if CLI flags can be set up quickly, do that
- generate a random table
- output it using glazed

## Examples

### Output formats

- json
- yaml
- csv
- ascii
- markdown
- html

### File output

- Single file output
- Multi file output

### Flattening structures

- json to rows

### Filtering columns

- filters and fields

### Go template support

- single string template
- multi file templates
- field templates

### Markdown output and templating

- Multi markdown template output with index page
- markdown template file

### Configuration file

- some examples

### Schema documentation

- show how to output schema

## Using glaze as a library

### Middlewares

- ObjectMiddleware
- RowMiddleware
- TableMiddleware

### Formatters

- TableOutputFormatter
- CsvOutputFormatter

### Command Line Integration

- cobra integration
- viper integration
- calibrate from config files

### Schema documentation

- show how to load different schemas

## The glaze tool

### Installation

- install with go get

### Import formats

- json / json rows / multiple files
- yaml / multiple files
- csv
- cut / ascii
- sqlite / SQL
- binary parser

### Output flags

## General brainstorm

- documentation for each subsystem

## Future ideas

- [-] SQLite output 
- API server to render local data
- parquet format (and pandas? numpy?)
- do we want some kind of transformation DSL / configuration DSL to do
  more complicated things? Definitely not at first, before having the use case for it.
- glaze tool
  - import and parse binary data
  - pcap input
- excel export
  - annotate excel export with as much metadata as possible
- HTML frontend
- jq integration
- markdown rendering with glow
- table app that can hide/show/rename/reorder columns
- search engine  / autocompletion based on known schema
  - use query language to create hyperlinks in output
- hyperlinked schema definitions
- sparklines and other shenanigans
- collect metadata and event logs to what led to the creation of the data itself
- cloud / network API output forms, for example to store something in s3 or other databases
  - SQL
  - dynamodb
  - s3
- style aliases (like pretty=oneline for git)
  - maybe styles can also have additional parameters