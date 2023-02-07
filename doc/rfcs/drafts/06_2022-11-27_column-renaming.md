# Add support for renaming columns

Ticket: [#27 - Add rename column middleware](https://github.com/go-go-golems/glazed/issues/27)

## Overview

Besides the plain need to rename columns when the originally selected fields
don't match, this middleware could be used to rename the output of templating
or the results of flattening objects.

## Brainstorm

### Implementation as row oriented middleware

This would lend itself well to a row processing middleware to preserve not only efficiency, 
but would be necessary for other potential row processing middleware that relies on the renamed
columns. 

### Possibility for a final rename

It should be possible to specify that the renaming happen at the end of the processing,
before presenting the results to a human. This is to make it possible to use programming-oriented
columnNames say when using templates, but to replace them with richer strings (using for example 
characters like whitespace or {} ) when outputting for human consumption.

## Command line argument support

It should be possible to provide a list of renames from the command line, probably through either
a repeated `--rename` flag, or through a comma separated list. 

The rename list could also be provided through its own configuration file that just contains
a dictionary of the renames, to make it easier to share across command line calls.