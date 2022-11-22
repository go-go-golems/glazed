# RFC - Configuration file support for glaze / CLI applications

Issue: #17

## Changelog

### 2022-11-14 - manuel

- Created document

## Overview

We want to configure a full output setting from a config file to
make it easier to repeatedly use pretty structured data output.

## Requirements

- make it easy for the developer to register config file support
- make it easy for the user to load a config file
- scaffold configuration file / output current configuration as config file

## Design brainstorm

I think we could make it easy to provide a config file (YAML, for example),
and load its parts into the configuration structs described in [flag-helpers](01_2022-11-13_flag-helpers.md).

These loaders could then easily be registered and called by the host program, using the middlewares
structure shown in the flag document (in fact, it's pretty much the exact same structure).