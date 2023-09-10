---
Title: Understanding Command Usage Strings
Slug: usage-string
Short: |
  Explains how command argument usage is indicated within our application.
Topics:
- User Guide
Commands: []
Flags: []
IsTemplate: false
IsTopLevel: true
ShowPerDefault: false
SectionType: GeneralTopic
---

## Overview

Every command stems around a verb which acts as the main action the command is performing.
In contrast to parameter flags, which are preceded by `--` for `-`, arguments are 
passed as normal arguments. They can be interleaved with normal flags.

Arguments can be:
- required. These need to be provided by the user, or have a default value
- list arguments, which means that they gobble up the rest of the arguments

When parsing the arguments, the parser will try to match the command-line arguments to the command's arguments.
Once the arguments run out, the parser will try to match the remaining arguments to their default values.

Finally, leftover arguments are assigned to a potential list argument.

- Required arguments are always placed before optional arguments.
- Parameters accepting list inputs should not directly follow each other.

### Required Parameters

Required arguments are indicated by angle brackets `<>`.
These arguments must be specified for a command to run successfully. For example:

```
command <filename>
```

This indicates that the command requires a filename to be specified for it to run properly.

### Optional Parameters

Optional arguments can be identified by the square brackets `[]`.
These arguments may be skipped, and the command may still run successfully. For example:

```
command <filename> [directory]
```

Here the `directory` is optional.
If not provided, the command will still execute, but with certain default settings.

### List Parameters

Some commands may accept a list of inputs for certain arguments. This is symbolized by an ellipsis `...` following the argument.

```
command <filename> [tags...]
```

In this case, the `tags` argument can accept a list of values.

### Default Values

Parameters may come with default values. These can be identified by text following the format `default: value`. This means that if you do not provide a value for this argument, the system will use the default value.

```
command <filename> [directory (default: home)]
```

This command implies that if no directory is specified,
the system uses the `home` directory as a default.

