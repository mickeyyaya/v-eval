"""Tests for release_notes: the section for one version is found by its Keep a Changelog
heading, printed without the neighbouring sections or the link references, and a missing
version is an error rather than empty notes."""
import contextlib
import io
import tempfile
import unittest
from pathlib import Path

import release_notes

CHANGELOG = """# Changelog

All notable changes are documented here.

## [Unreleased]

Nothing yet.

## [0.2.0] - 2026-10-01

Second release.

### Added

- A thing.

## [0.1.0] - 2026-09-15

First release.

### Added

- The core.
- The command line.

[Unreleased]: https://example.org/compare/v0.2.0...HEAD
[0.2.0]: https://example.org/compare/v0.1.0...v0.2.0
[0.1.0]: https://example.org/releases/tag/v0.1.0
"""


def run_main(argv: list[str]) -> tuple[int, str, str]:
    out, err = io.StringIO(), io.StringIO()
    with contextlib.redirect_stdout(out), contextlib.redirect_stderr(err):
        code = release_notes.main(argv)
    return code, out.getvalue(), err.getvalue()


class SectionTest(unittest.TestCase):
    def test_dated_heading_yields_the_body_up_to_the_next_section(self) -> None:
        self.assertEqual(release_notes.section(CHANGELOG, "0.2.0"), "Second release.\n\n### Added\n\n- A thing.\n")

    def test_plain_heading_forms_are_accepted(self) -> None:
        bracketed = "## [0.1.0]\n\nBracketed.\n\n## [0.0.9]\n\nOlder.\n"
        bare = "## 0.1.0\n\nBare.\n\n## 0.0.9\n\nOlder.\n"
        self.assertEqual(release_notes.section(bracketed, "0.1.0"), "Bracketed.\n")
        self.assertEqual(release_notes.section(bare, "0.1.0"), "Bare.\n")

    def test_version_must_match_the_whole_token(self) -> None:
        text = "## [0.1.0-rc1] - 2026-09-01\n\nCandidate.\n\n## [10.1.0] - 2026-09-02\n\nTen.\n"
        self.assertIsNone(release_notes.section(text, "0.1.0"))
        self.assertEqual(release_notes.section(text, "10.1.0"), "Ten.\n")

    def test_missing_version_is_none(self) -> None:
        self.assertIsNone(release_notes.section(CHANGELOG, "9.9.9"))

    def test_last_section_stops_before_the_link_references(self) -> None:
        body = release_notes.section(CHANGELOG, "0.1.0")
        self.assertEqual(body, "First release.\n\n### Added\n\n- The core.\n- The command line.\n")

    def test_headings_inside_fenced_code_are_not_section_boundaries(self) -> None:
        text = (
            "## [0.2.0] - 2026-10-01\n\nShows a heading:\n\n```markdown\n## [0.1.0] - 1999-01-01\n```\n\nStill 0.2.0.\n\n"
            "## [0.1.0] - 2026-09-15\n\nReal 0.1.0.\n"
        )
        self.assertEqual(
            release_notes.section(text, "0.2.0"),
            "Shows a heading:\n\n```markdown\n## [0.1.0] - 1999-01-01\n```\n\nStill 0.2.0.\n",
        )
        self.assertEqual(release_notes.section(text, "0.1.0"), "Real 0.1.0.\n")

    def test_link_references_inside_fenced_code_do_not_end_the_section(self) -> None:
        text = "## [0.1.0]\n\n```text\n[ref]: https://example.org\n```\n\nAfter.\n\n[0.1.0]: https://example.org/tag\n"
        self.assertEqual(release_notes.section(text, "0.1.0"), "```text\n[ref]: https://example.org\n```\n\nAfter.\n")

    def test_body_is_trimmed_to_one_trailing_newline(self) -> None:
        text = "## [0.1.0]\n\n\n\nBody.\n\n\n\n## [0.0.9]\n"
        self.assertEqual(release_notes.section(text, "0.1.0"), "Body.\n")


class ProseLinesTest(unittest.TestCase):
    def test_tilde_fence_hides_lines_inside_from_being_prose(self) -> None:
        text = "~~~\n## not a heading\n~~~"
        self.assertEqual([prose for _, prose in release_notes.prose_lines(text)], [False, False, False])

    def test_unterminated_fence_runs_to_the_end_of_the_document(self) -> None:
        text = "```\n## not a heading\nstill fenced"
        self.assertEqual([prose for _, prose in release_notes.prose_lines(text)], [False, False, False])


class MainTest(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.root = Path(self.tmp.name)
        self.changelog = self.root / "CHANGELOG.md"
        self.changelog.write_text(CHANGELOG, encoding="utf-8")

    def tearDown(self) -> None:
        self.tmp.cleanup()

    def test_found_section_is_printed_and_exits_0(self) -> None:
        code, out, err = run_main([str(self.changelog), "0.2.0"])
        self.assertEqual((code, out, err), (0, "Second release.\n\n### Added\n\n- A thing.\n", ""))

    def test_missing_version_exits_2_with_a_message(self) -> None:
        code, out, err = run_main([str(self.changelog), "9.9.9"])
        self.assertEqual((code, out), (2, ""))
        self.assertEqual(err, f"error: no section for version 9.9.9 in {self.changelog}\n")

    def test_v_prefixed_version_exits_2_with_the_no_section_message(self) -> None:
        code, out, err = run_main([str(self.changelog), "v0.1.0"])
        self.assertEqual((code, out), (2, ""))
        self.assertEqual(err, f"error: no section for version v0.1.0 in {self.changelog}\n")

    def test_empty_section_exits_2_and_writes_nothing(self) -> None:
        empty = self.root / "empty.md"
        empty.write_text("## [0.1.0] - 2026-09-15\n\n## [0.0.9]\n", encoding="utf-8")
        target = self.root / "notes.md"
        code, out, err = run_main([str(empty), "0.1.0", "-o", str(target)])
        self.assertEqual((code, out), (2, ""))
        self.assertEqual(err, f"error: section for version 0.1.0 in {empty} is empty\n")
        self.assertFalse(target.exists())

    def test_blank_only_section_exits_2_with_a_message(self) -> None:
        blank = self.root / "blank.md"
        blank.write_text("## [0.1.0] - 2026-09-15\n\n\n\n## [0.0.9]\n", encoding="utf-8")
        code, out, err = run_main([str(blank), "0.1.0"])
        self.assertEqual((code, out), (2, ""))
        self.assertEqual(err, f"error: section for version 0.1.0 in {blank} is empty\n")

    def test_unreadable_input_exits_2(self) -> None:
        missing = self.root / "missing.md"
        code, out, err = run_main([str(missing), "0.1.0"])
        self.assertEqual((code, out), (2, ""))
        self.assertTrue(err.startswith(f"error: cannot read {missing}:"), err)

    def test_output_file_is_written_with_lf_and_nothing_is_printed(self) -> None:
        target = self.root / "build" / "release-notes.md"
        code, out, err = run_main([str(self.changelog), "0.1.0", "-o", str(target)])
        self.assertEqual((code, out, err), (0, "", ""))
        self.assertEqual(target.read_bytes(), b"First release.\n\n### Added\n\n- The core.\n- The command line.\n")

    def test_crlf_input_yields_lf_output(self) -> None:
        self.changelog.write_bytes(CHANGELOG.replace("\n", "\r\n").encode("utf-8"))
        target = self.root / "notes.md"
        code, _, _ = run_main([str(self.changelog), "0.1.0", "-o", str(target)])
        self.assertEqual(code, 0)
        self.assertEqual(target.read_bytes(), b"First release.\n\n### Added\n\n- The core.\n- The command line.\n")


if __name__ == "__main__":
    unittest.main()
