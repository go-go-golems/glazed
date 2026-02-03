#!/usr/bin/env python3
import re
from pathlib import Path

ROOT = Path('.')
SKIP_DIRS = {'.git', 'node_modules', 'vendor', 'ttmp'}

REPLACEMENTS = [
    (r"\bParsedParametersFromDefaults\b", "FieldValuesFromDefaults"),
    (r"\bNewParsedParameters\b", "NewFieldValues"),
    (r"\bParsedParametersOption\b", "FieldValuesOption"),
    (r"\bParsedParameters\b", "FieldValues"),
    (r"\bParsedParameter\b", "FieldValue"),
    (r"\bWithParsedParameter\b", "WithFieldValue"),
    (r"\bToSerializableParsedParameters\b", "ToSerializableFieldValues"),
    (r"\bToSerializableParsedParameter\b", "ToSerializableFieldValue"),
    (r"\bSerializableParsedParameters\b", "SerializableFieldValues"),
    (r"\bSerializableParsedParameter\b", "SerializableFieldValue"),
    (r"\bGatherParametersFromMap\b", "GatherFieldsFromMap"),
    (r"\bParseParameter\b", "ParseField"),
    (r"\bCheckParameterDefaultValueValidity\b", "CheckDefaultValueValidity"),
    (r"\bAddParametersToCobraCommand\b", "AddFieldsToCobraCommand"),
    (r"\bparsedParameters\b", "fieldValues"),
    (r"\bparsedParameter\b", "fieldValue"),
]

FILE_GLOBS = ["**/*.go"]

compiled = [(re.compile(pat), repl) for pat, repl in REPLACEMENTS]

changed_files = []

for glob in FILE_GLOBS:
    for path in ROOT.glob(glob):
        if not path.is_file():
            continue
        if any(part in SKIP_DIRS for part in path.parts):
            continue
        text = path.read_text()
        new_text = text
        for rx, repl in compiled:
            new_text = rx.sub(repl, new_text)
        if new_text != text:
            path.write_text(new_text)
            changed_files.append(path)

print(f"Updated {len(changed_files)} files")
