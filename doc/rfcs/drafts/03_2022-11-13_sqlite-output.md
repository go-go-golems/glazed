# RFC - Add SQLite output

Issue: #23

## Changelog

### 2022-11-13 - manuel

- Created document
- Gathered initial requirements
- Brainstorm

## Overview

We want to add SQLite output support.

## Brainstorm

### 2022-11-13

- how do we create the schema?
- do we want to handle multiple tables?
- what about foreign keys
- what command line arguments for glaze can we give to split a table into multiple?
- how portable do we build the SQL output so that we can easily hook up other databases?
  - use a sqlbuilder (!)