#!/usr/bin/env bash

instruct() {
	echo "<taskInstructions>"
	echo "$1"
	echo "</taskInstructions>"
	echo

}

INSTRUCTIONS="Write a documentation entry and then also write a tutorial and the nwrite a cool application about how to use Commands from glazed. 

The base package is github.com/go-go-golems/glazed.

Topics you should mention: 

- creating a command
- adding flags/arguments to commands
- how to run a command (for all the different Run methods)
- how to use a loader to load commands from yaml

Use the help format described.
"

instruct "$INSTRUCTIONS"

echo
echo "---"
echo

instruct "Here is some existing documentation about using commands:"

echo "<information>"
prompto get glazed/create-command  -- --cobra 
echo "</information>"

instruct "Here is how parameters for commands work"

echo "<information>"
prompto get glazed/parameters-verbose
echo "</information>"

echo
echo "---"
echo
instruct "Here is the source code of the Command system"

echo "<information>"
catter pkg/cmds/cmds.go pkg/cmds/loaders pkg/cmds/alias
echo "</information>"

echo
echo "---"
echo

instruct "This is the help format you should use:"

echo "<information>"
prompto get glazed/writing-help-entries
echo "</information>"

echo "<example type='tutorial'>"
catter pkg/doc/tutorials/01-a-simple-table-cli.md
echo "</example>"

echo "<example type='topic'>"
catter pkg/doc/topics/02-markdown-style.md
echo "</example>"

echo "<example type='application'>"
catter pkg/doc/applications/01-exposing-a-simple-sql-table.md
echo "</example>"

echo
echo "---"
echo

instruct "Here are your instructions again:"

instruct "$INSTRUCTIONS"
