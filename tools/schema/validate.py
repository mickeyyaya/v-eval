#!/usr/bin/env python3
"""Validate JSON instance files against one JSON Schema.

``python tools/schema/validate.py SCHEMA INSTANCE...`` prints one line per instance -- ``ok``,
or how many errors it has -- followed by the first of those errors (``--max-errors``, default
3), each with the JSON path it was found at. Exits 0 when every instance is valid, 1 when any
is not, and 2 when the check could not run: a file that cannot be read or parsed, a schema that
is not itself valid, or the ``jsonschema`` package not installed. The schema's own ``$schema``
picks the draft; ``format`` annotations are not asserted.

Continuous-integration dependency only (decision 0022, amended 2026-09-15): this is the one tool
in the repository that imports a third-party package, ``jsonschema``, because no JSON Schema
validator exists in the standard library. CI installs it (``.github/workflows/go.yml``, job
``schemas``); a maintainer installs it on demand with ``pip install jsonschema``. The product and
every other tool stay standard-library. Runs on macOS, Linux, and Windows.
"""
from __future__ import annotations

import argparse
import json
import sys
from typing import Any, Sequence

MISSING_PACKAGE = "jsonschema is not installed (pip install jsonschema)"


class ToolError(Exception):
    """A user-facing failure: reported on stderr with exit code 2."""


def read_json(path: str) -> Any:
    try:
        with open(path, encoding="utf-8") as handle:
            return json.load(handle)
    except UnicodeDecodeError as exc:
        raise ToolError(f"cannot read {path}: not valid UTF-8 ({exc.reason} at byte {exc.start})") from exc
    except OSError as exc:
        raise ToolError(f"cannot read {path}: {exc.strerror}") from exc
    except ValueError as exc:
        raise ToolError(f"cannot parse {path}: {exc}") from exc


def load_jsonschema() -> Any:
    """The jsonschema package, imported only once a schema has been read, so that every failure
    that needs no validator is reported without it."""
    try:
        import jsonschema  # noqa: PLC0415 -- deliberately deferred; see the module docstring
    except ModuleNotFoundError as exc:
        if exc.name != "jsonschema":  # a broken dependency inside the package is a different problem
            raise
        raise ToolError(MISSING_PACKAGE) from exc
    return jsonschema


def build_validator(jsonschema: Any, schema: Any, path: str) -> Any:
    """A validator for the draft the schema declares, after checking the schema itself."""
    cls = jsonschema.validators.validator_for(schema)
    try:
        cls.check_schema(schema)
    except jsonschema.exceptions.SchemaError as exc:
        raise ToolError(f"{path} is not a valid schema: {exc.message}") from exc
    return cls(schema)


def errors_for(validator: Any, instance: Any) -> list[str]:
    """Every error the validator finds, as ``<json path>: <message>``, in path order."""
    found = sorted(validator.iter_errors(instance), key=lambda error: error.json_path)
    return [f"{error.json_path}: {error.message}" for error in found]


def report(path: str, errors: list[str], max_errors: int) -> None:
    if not errors:
        print(f"{path}: ok")
        return
    noun = "error" if len(errors) == 1 else "errors"
    print(f"{path}: {len(errors)} {noun}")
    for error in errors[:max_errors]:
        print(f"  {error}")


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="validate", description=__doc__.split("\n\n")[0])
    parser.add_argument("schema", help="path to the JSON Schema")
    parser.add_argument("instances", nargs="+", metavar="INSTANCE", help="JSON files to validate against it")
    parser.add_argument("--max-errors", type=int, default=3, metavar="N", help="print at most N errors per instance (default 3)")
    return parser


def main(argv: Sequence[str] | None = None) -> int:
    args = build_parser().parse_args(argv)
    invalid = 0
    try:
        schema = read_json(args.schema)
        validator = build_validator(load_jsonschema(), schema, args.schema)
        for path in args.instances:
            errors = errors_for(validator, read_json(path))
            invalid += bool(errors)
            report(path, errors, args.max_errors)
    except ToolError as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 2
    return 1 if invalid else 0


if __name__ == "__main__":
    raise SystemExit(main())
