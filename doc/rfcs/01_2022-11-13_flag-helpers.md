# RFC - Add helpers to help CLI writers handle command line arguments

## Changelog

### 2022-11-13 - manuel

- Created document
- Gathered initial requirements
- Brainstorm

## Overview

The goal is to make it as easy as possible for developers writing CLI tools
to give command line flags and configuration file options to control
glaze output.

## Requirements

- support cobra
- support standard golang flags
- make it easy to use with other flag frameworks

There are two complicated aspects:

- one is making it easy to generate nice documentation examples
- one is to configure the help strings with options that are relevant to the domain

One option to create nice configuration examples would be to give the host program the option
to pass in synthetic data along with a description of the flag values set.

## Design brainstorm

The following entries are ideas

### 2022-11-13

We could have one structure to describe command line flags. A collection
of these flag structures is attached to the output system, middleware and other
configurable component or subsystem.

The flag structures can then be used to configure the different ways of providing configuration
(and make it easy for users to define new adapters). These structures can then be used to instantiate
the necessary structures.

This could look something like:

```go
package cli

type Flag struct {
	Name             string
	LongOption       string
	ShortOption      string
	Type             string
	ShortDescription string
	LongDescription  string
}

// potentially we could imagine a structure to do nested configurations for more complex subsystems,
// for example for go field templates. Although to be honest, the value is more in being able to register
// and parse flags in a generic manner.

type OutputConfigurationFlags struct {
	Output       Flag
	OutputFile   Flag
	TableFormat  Flag
	WithHeaders  Flag
	CSVSeparator Flag
}

type OutputConfiguration struct {
	Output       string
	OutputFile   string
	TableFormat  string
	WithHeaders  bool
	CSVSeparator string
}
```

It would then be possible for a developer to get the `OutputConfigurationFlags`,
use that populate their config file / command line flag structure of choice,
and ultimately to create a `OutputConfiguration` and pass that to the entry point of the `output`
subsystem.

Note that we haven't tackled how to validate these values, how to provide autocompletion hints,
nor how to parse command line flags into more complex structures (see go template fields).

I think these systems should be super easy, deal only with flags and single strings.
Anything more complicated is either something we provide for easy hookup too, 
or is something that can be written by the user in more detail.



