# RFC - Output data into more than one file

## Changelog

### 2022-11-13 - manuel

- Created document
- Wrote up some potential use cases

## Overview

Structured data can often be presented best when split into multiple files. 

## Design brainstorm

### 2022-11-13

Some use case examples:

- one file per row in the original data, rendered in markdown (think, recipe book)
- split out each objects in a JSON list into its own file
- render schema documentation along with differently filtered tables
- split a table into multiple table based on a category column
- create an overview file that links to other data files
- annotate output data with metadata (event logs, creation metadata, signatures...)
