"""Tests for validate: the command line, its exit codes, and its messages. None of them needs
the jsonschema package: the missing-package path is tested with the import mocked absent, and
the two tests that validate real instances are skipped, with a reason, when the package is not
importable, so the suite runs on a bare interpreter in CI on every operating system."""
import contextlib
import importlib.util
import io
import json
import sys
import tempfile
import unittest
from pathlib import Path
from unittest import mock

import validate

HAS_JSONSCHEMA = importlib.util.find_spec("jsonschema") is not None

SCHEMA = {
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "type": "object",
    "required": ["name"],
    "properties": {"name": {"type": "string"}},
    "additionalProperties": False,
}


def run_main(argv: list[str]) -> tuple[int, str, str]:
    out, err = io.StringIO(), io.StringIO()
    with contextlib.redirect_stdout(out), contextlib.redirect_stderr(err):
        code = validate.main(argv)
    return code, out.getvalue(), err.getvalue()


class FilesMixin(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.root = Path(self.tmp.name)
        self.schema = self.write("schema.json", SCHEMA)
        self.valid = self.write("valid.json", {"name": "x"})
        self.invalid = self.write("invalid.json", {"name": 1, "extra": True})

    def tearDown(self) -> None:
        self.tmp.cleanup()

    def write(self, name: str, value: object) -> Path:
        path = self.root / name
        path.write_text(json.dumps(value), encoding="utf-8")
        return path


class WithoutJSONSchemaTest(FilesMixin):
    def test_missing_package_exits_2_with_the_install_hint(self) -> None:
        with mock.patch.dict(sys.modules, {"jsonschema": None}):
            code, out, err = run_main([str(self.schema), str(self.valid)])
        self.assertEqual((code, out), (2, ""))
        self.assertEqual(err, "error: jsonschema is not installed (pip install jsonschema)\n")

    def test_unreadable_schema_exits_2_before_the_package_is_needed(self) -> None:
        missing = self.root / "missing.json"
        with mock.patch.dict(sys.modules, {"jsonschema": None}):
            code, out, err = run_main([str(missing), str(self.valid)])
        self.assertEqual((code, out), (2, ""))
        self.assertTrue(err.startswith(f"error: cannot read {missing}:"), err)

    def test_schema_that_is_not_json_exits_2(self) -> None:
        broken = self.root / "broken.json"
        broken.write_text("{not json", encoding="utf-8")
        with mock.patch.dict(sys.modules, {"jsonschema": None}):
            code, out, err = run_main([str(broken), str(self.valid)])
        self.assertEqual((code, out), (2, ""))
        self.assertTrue(err.startswith(f"error: cannot parse {broken}:"), err)


@unittest.skipUnless(HAS_JSONSCHEMA, "jsonschema is not installed")
class WithJSONSchemaTest(FilesMixin):
    def test_valid_instances_exit_0_and_say_ok(self) -> None:
        code, out, err = run_main([str(self.schema), str(self.valid), str(self.valid)])
        self.assertEqual((code, err), (0, ""))
        self.assertEqual(out, f"{self.valid}: ok\n{self.valid}: ok\n")

    def test_invalid_instance_exits_1_with_the_count_and_the_first_errors(self) -> None:
        code, out, err = run_main([str(self.schema), str(self.valid), str(self.invalid)])
        self.assertEqual((code, err), (1, ""))
        lines = out.splitlines()
        self.assertEqual(lines[0], f"{self.valid}: ok")
        self.assertEqual(lines[1], f"{self.invalid}: 2 errors")
        self.assertEqual(len(lines), 4)
        self.assertTrue(lines[2].startswith("  $: Additional properties are not allowed"), lines[2])
        self.assertEqual(lines[3], "  $.name: 1 is not of type 'string'")

    def test_max_errors_caps_the_lines_but_not_the_count(self) -> None:
        code, out, _ = run_main(["--max-errors", "1", str(self.schema), str(self.invalid)])
        self.assertEqual(code, 1)
        self.assertEqual(out, f"{self.invalid}: 2 errors\n  $: Additional properties are not allowed ('extra' was unexpected)\n")

    def test_an_invalid_schema_is_a_tool_error(self) -> None:
        bad = self.write("bad-schema.json", {"type": "no-such-type"})
        code, out, err = run_main([str(bad), str(self.valid)])
        self.assertEqual((code, out), (2, ""))
        self.assertTrue(err.startswith(f"error: {bad} is not a valid schema:"), err)

    def test_an_instance_that_is_not_json_is_a_tool_error(self) -> None:
        broken = self.root / "broken.json"
        broken.write_text("[", encoding="utf-8")
        code, out, err = run_main([str(self.schema), str(self.valid), str(broken)])
        self.assertEqual(code, 2)
        self.assertEqual(out, f"{self.valid}: ok\n")
        self.assertTrue(err.startswith(f"error: cannot parse {broken}:"), err)


if __name__ == "__main__":
    unittest.main()
