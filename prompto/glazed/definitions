#!/usr/bin/env bash

for i in CommandDescription ParameterDefinition ParameterLayer; do
	oak go definitions --recurse pkg/ --name "$i" --definition-type struct,interface
done
