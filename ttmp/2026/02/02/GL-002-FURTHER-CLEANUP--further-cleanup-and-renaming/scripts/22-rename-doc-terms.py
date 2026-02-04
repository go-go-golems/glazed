#!/usr/bin/env python3
import re
import pathlib

ROOT = pathlib.Path(".")

TARGET_GLOBS = [
    "README.md",
    "**/*.md",
    "**/*.yaml",
    "**/*.yml",
    "**/*.txt",
    "**/*.json",
    "**/*.prompto",
    "prompto/**",
    "pinocchio/**",
]

EXCLUDE_DIRS = {"ttmp", ".git", "node_modules", "vendor"}

REPLACEMENTS = [
    # File/path references
    ("pkg/cmds/layers", "pkg/cmds/schema"),
    ("pkg/cmds/parameters", "pkg/cmds/fields"),
    ("layers-guide", "sections-guide"),
    ("logging-layer", "logging-section"),
    ("custom-layer", "custom-section"),
    ("parameter-types", "field-types"),
    ("parameters-verbose", "fields-verbose"),
    ("parameters", "fields"),
    # Config keys / identifiers
    ("target_layer", "target_section"),
    ("target_parameter", "target_field"),
    ("layerSlug", "sectionSlug"),
    ("layerSlugs", "sectionSlugs"),
    ("layerName", "sectionName"),
    ("layer_name", "section_name"),
    ("layer_params", "section_fields"),
    ("param_name", "field_name"),
    ("param_info", "field_info"),
    ("layers_", "schema_"),
    ("layersToMerge", "sectionsToMerge"),
    ("layerToMerge", "sectionToMerge"),
    ("parameterSections", "schema_"),
    # Fully-qualified type replacements
    ("layers.ParsedLayers", "values.Values"),
    ("layers.ParsedLayer", "values.SectionValues"),
    ("layers.ParameterLayers", "schema.Schema"),
    ("layers.ParameterLayer", "schema.Section"),
    ("layers.DefaultSlug", "schema.DefaultSlug"),
    ("parameters.ParsedParameters", "fields.FieldValues"),
    ("parameters.ParsedParameter", "fields.FieldValue"),
    ("parameters.ParameterDefinitions", "fields.Definitions"),
    ("parameters.ParameterDefinition", "fields.Definition"),
    ("parameters.ParameterType", "fields.Type"),
    # Unqualified type replacements
    ("ParsedLayers", "Values"),
    ("ParsedLayer", "SectionValues"),
    ("ParameterLayers", "Schema"),
    ("ParameterLayer", "Section"),
    ("ParsedParameters", "FieldValues"),
    ("ParsedParameter", "FieldValue"),
    ("ParameterDefinitions", "Definitions"),
    ("ParameterDefinition", "Definition"),
    ("ParameterType", "Type"),
    # CamelCase token replacements
    ("Layers", "Sections"),
    ("Layer", "Section"),
    ("Parameters", "Fields"),
    ("Parameter", "Field"),
]

WORD_REPLACEMENTS = [
    (re.compile(r"\blayers\b"), "sections"),
    (re.compile(r"\blayer\b"), "section"),
    (re.compile(r"\bparameters\b"), "fields"),
    (re.compile(r"\bparameter\b"), "field"),
]


def should_skip(path: pathlib.Path) -> bool:
    parts = set(path.parts)
    if parts & EXCLUDE_DIRS:
        return True
    return False


def iter_files():
    seen = set()
    for pattern in TARGET_GLOBS:
        for path in ROOT.glob(pattern):
            if path.is_dir() or path in seen:
                continue
            if should_skip(path):
                continue
            seen.add(path)
            yield path


def rewrite_text(text: str) -> str:
    for old, new in REPLACEMENTS:
        text = text.replace(old, new)
    for regex, repl in WORD_REPLACEMENTS:
        text = regex.sub(repl, text)
    return text


def main() -> None:
    for path in iter_files():
        try:
            data = path.read_text()
        except Exception:
            continue
        new_data = rewrite_text(data)
        if new_data != data:
            path.write_text(new_data)


if __name__ == "__main__":
    main()
