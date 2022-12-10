---
Title: Help System
Slug: help-system
Short: glazed comes with a powerful help system to make it easy to create rich CLI help pages.
Topics:
- help
Commands:
- help
Flags:
- flag
- topic
- command
- list
- topics
- examples
- applications
- tutorials
- help
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

## Overview

In addition to the command system provided by your CLI flag parsing handler
(like cobra), glazed allows you to augment that output by providing 4 different
types of sections:

- general topics sections, which you can think of as general articles
- example sections, which show case one specific way of using a command or a flag
- application sections, which show interesting ways of using some functionality, often using additional external programs
- tutorials, which are step by step guides for using a specific functionality

Multiple sections get combined to create one help page (`GenericHelpPage` in the
library), which is then handed off to various go templates, which in turn create 
a markdown output.

TODO(manuel, 2022-12-09) - we could use a nice ASCII diagram of the pipeline here

## Using the `help` command

(This is currently only implemented for `cobra` applications, for example for `glaze`).

The help system is accessible through the cobra help system.

You can use `glaze help <slug>` to access any section from the command line.
This can be used to display a topic page (which will also show related examples,
applications or tutorials), or a single section in full.

To get help on the options of the `help` command itself, you can run :

``` 
❯ glaze help help

   help - Help about any command or topic                                                              
                                                                                                       
  Help provides help for any command and topic in the application.                                     
                                                                                                       
  For more help, run:  glaze help help                                                                 
                                                                                                       
  ## Usage:                                                                                            
                                                                                                       
   glaze help [topic/command] [flags]                                                                  
                                                                                                       
  ## Flags:                                                                                            
                                                                                                       
          --all              Show all sections, not just default                                       
          --applications     Show all applications                                                     
          --command string   Show help related to command                                              
          --examples         Show all examples                                                         
          --flag string      Show help related to flag                                                 
      -h, --help             help for help                                                             
          --list             List all sections                                                         
          --short            Show short version                                                        
          --topic string     Show help related to topic                                                
          --topics           Show all topics                                                           
          --tutorials        Show all tutorials                                                        
                                                                                                       
  ## Help System:                                                                                      
                                                                                                       
  glazed comes with a powerful help system to make it easy to create rich CLI help pages.              
                                                                                                       
  To learn more, run:  glaze help help-system                                                          

```

To get an overview of the toplevel help sections, you can run:
```

❯ glaze help --list

   glaze - glaze is a tool to format structured data                                                   
                                                                                                       
  For more help, run:  glaze help glaze                                                                
                                                                                                       
  ## General topics                                                                                    
                                                                                                       
  Run  glaze help <topic>  to view a topic's page.                                                     
                                                                                                       
  • help-system - Help System                                                                          
  • templates - Templates                                                                              
                                                                                                       
  ## Examples                                                                                          
                                                                                                       
  Run  glaze help <example>  to view an example in full.                                               
                                                                                                       
  • templates-example-1 - Use a single template for single field output                                

```

## Section structure

Each section has:
- a `Title`
- a `SubTitle`
- a `Short` description 
- a full `Content`

The short description and full content are plain markdown.
For examples, the short description should be the full command line of the example,
potentially over multiple lines.

The `Slug` is similar to an id and used to reference the section internally.

Furthermore, each section has a list of "topics" (slugs of other help sections
that it is related to), flags and commands that are relevant.
This metadata is used to find the sections to be shown when a user
requests the help for a command, or for a flag, or to show related topics.

## Default sections

Some sections are shown by default. For example, when calling up the help for a command,
the general topics,examples, applications and tutorials related to that command and that
have the `ShowPerDefault` flag will be shown without further flags.

Sections that don't have the `ShowPerDefault` flag set however will only be shown when
explicitly asked for using the `--topics` `--flags` `--examples` options.

## Querying help pages

glazed augments the help system by augmenting each help output (say, when 
running `command --help`) with its related pages.

For example, if the user requests the help for the `json` command,
glazed will look for all the sections related to the `json` command (in the `Commands` 
metadata entry), sort them into a `GenericHelpPage`, and then render them using
one of the `help-long-section-list.tmpl` or `help-short-section-list.tmpl` templates.

TODO(manuel, 2022-12-09): Add more information about how we actually query sections (using SectionQuery)

## Creating help pages using go embed

These pages are most easily included into the CLI utility by loading them
from markdown files at compile time using the `go:embed` functionality.
