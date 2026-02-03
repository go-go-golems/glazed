#!/usr/bin/env python3
import pathlib

PATH = pathlib.Path("cmd/examples/parameter-types/main.go")

IDENT_RENAMES = {
    "StringListFromFilesParam": "StringListFromFilesField",
    "StringListFromFileParam": "StringListFromFileField",
    "StringFromFilesParam": "StringFromFilesField",
    "StringFromFileParam": "StringFromFileField",
    "ObjectListFromFilesParam": "ObjectListFromFilesField",
    "ObjectListFromFileParam": "ObjectListFromFileField",
    "ObjectFromFileParam": "ObjectFromFileField",
    "IntegerListParam": "IntegerListField",
    "StringListParam": "StringListField",
    "FloatListParam": "FloatListField",
    "ChoiceListParam": "ChoiceListField",
    "FileListParam": "FileListField",
    "KeyValueParam": "KeyValueField",
    "StringParam": "StringField",
    "SecretParam": "SecretField",
    "IntegerParam": "IntegerField",
    "FloatParam": "FloatField",
    "BoolParam": "BoolField",
    "DateParam": "DateField",
    "ChoiceParam": "ChoiceField",
    "FileParam": "FileField",
}

TEXT_RENAMES = {
    "parameter-types": "field-types",
    "Parameter types": "Field types",
    "parameter types": "field types",
    "Parameter type": "Field type",
    "parameter type": "field type",
    "parameters": "fields",
    "parameter": "field",
}

TAG_RENAMES = {
    "string-param": "string-field",
    "secret-param": "secret-field",
    "integer-param": "integer-field",
    "float-param": "float-field",
    "bool-param": "bool-field",
    "date-param": "date-field",
    "choice-param": "choice-field",
    "string-list-param": "string-list-field",
    "integer-list-param": "integer-list-field",
    "float-list-param": "float-list-field",
    "choice-list-param": "choice-list-field",
    "file-param": "file-field",
    "file-list-param": "file-list-field",
    "string-from-file-param": "string-from-file-field",
    "string-from-files-param": "string-from-files-field",
    "string-list-from-file-param": "string-list-from-file-field",
    "string-list-from-files-param": "string-list-from-files-field",
    "object-from-file-param": "object-from-file-field",
    "object-list-from-file-param": "object-list-from-file-field",
    "object-list-from-files-param": "object-list-from-files-field",
    "key-value-param": "key-value-field",
}


def main() -> None:
    data = PATH.read_text()

    # Identifiers (longest first to avoid partial overlaps)
    for old in sorted(IDENT_RENAMES, key=len, reverse=True):
        data = data.replace(old, IDENT_RENAMES[old])

    # Tag and string literals (specific before generic)
    for old in sorted(TAG_RENAMES, key=len, reverse=True):
        data = data.replace(old, TAG_RENAMES[old])

    for old in sorted(TEXT_RENAMES, key=len, reverse=True):
        data = data.replace(old, TEXT_RENAMES[old])

    PATH.write_text(data)


if __name__ == "__main__":
    main()
