"""Tests for check_links: relative links and heading anchors must resolve, external
links are ignored, and unusual cases are reported rather than silently passed."""
import os
import tempfile
import unittest
from pathlib import Path

import check_links


def write(root: Path, rel: str, text: str) -> None:
    path = root / rel
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(text, encoding="utf-8")


class EdgeCaseTest(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.root = Path(self.tmp.name)

    def tearDown(self) -> None:
        self.tmp.cleanup()

    def test_angle_bracketed_target_with_spaces_is_checked(self) -> None:
        write(self.root, "a.md", "[doc](<my file.md>) [gone](<no file.md>)")
        write(self.root, "my file.md", "# M\n")
        broken = check_links.find_broken(str(self.root))
        self.assertEqual(broken, [check_links.BrokenLink("a.md", "no file.md", "missing file")])

    def test_unreadable_file_is_a_tool_error_with_the_path(self) -> None:
        (self.root / "bad.md").write_bytes(b"\xff\xfe not utf-8")
        with self.assertRaises(check_links.ToolError) as ctx:
            check_links.find_broken(str(self.root))
        self.assertIn("bad.md", str(ctx.exception))

    def test_target_with_balanced_parentheses_resolves(self) -> None:
        write(self.root, "a.md", "[link](report_(v2).md) and [Wiki](https://en.wikipedia.org/wiki/Bracket_(mathematics))")
        write(self.root, "report_(v2).md", "# R\n")
        self.assertEqual(check_links.find_broken(str(self.root)), [])

    def test_case_mismatch_is_missing_on_every_platform(self) -> None:
        write(self.root, "README.md", "# Top\n")
        write(self.root, "a.md", "[bad case](readme.md)")
        self.assertEqual(check_links.find_broken(str(self.root)), [check_links.BrokenLink("a.md", "readme.md", "missing file")])

    def test_link_escaping_the_root_is_missing(self) -> None:
        write(self.root, "a.md", "[outside](../etc/passwd)")
        self.assertEqual(check_links.find_broken(str(self.root)), [check_links.BrokenLink("a.md", "../etc/passwd", "missing file")])

    def test_unterminated_fence_hides_links_to_end_of_document(self) -> None:
        write(self.root, "a.md", "# A\n\n```text\n[fake](nope.md)\nnever closed\n")
        self.assertEqual(check_links.find_broken(str(self.root)), [])

    def test_longer_closing_fence_ends_the_block(self) -> None:
        write(self.root, "a.md", "```\n[fake](nope.md)\n`````\n[real](b.md)")
        write(self.root, "b.md", "# B\n")
        self.assertEqual(check_links.find_broken(str(self.root)), [])

    @unittest.expectedFailure
    def test_fence_indented_inside_a_list_item_is_still_code(self) -> None:
        """Known limitation: fences indented four or more columns inside list items are not recognized."""
        write(self.root, "a.md", "- item\n\n    ```\n    [fake](nope.md)\n    ```\n")
        self.assertEqual(check_links.find_broken(str(self.root)), [])

    def test_absolute_path_target_is_reported_as_not_checkable(self) -> None:
        write(self.root, "etc/passwd.md", "# P\n")
        write(self.root, "a.md", "[abs](/etc/passwd.md)")
        self.assertEqual(check_links.find_broken(str(self.root)), [check_links.BrokenLink("a.md", "/etc/passwd.md", check_links.ABSOLUTE_TARGET)])

    def test_link_to_a_directory_resolves(self) -> None:
        write(self.root, "docs/a.md", "# A\n")
        write(self.root, "README.md", "[docs](docs/) [docs2](docs)")
        self.assertEqual(check_links.find_broken(str(self.root)), [])

    def test_dangling_symlink_is_a_tool_error_not_a_crash(self) -> None:
        try:
            os.symlink(str(self.root / "missing.md"), str(self.root / "dangling.md"))
        except (OSError, NotImplementedError):
            self.skipTest("symlinks not permitted here")
        with self.assertRaises(check_links.ToolError) as ctx:
            check_links.find_broken(str(self.root))
        self.assertIn("dangling.md", str(ctx.exception))

    def test_reported_paths_use_forward_slashes_on_every_platform(self) -> None:
        write(self.root, "docs/a.md", "[gone](../missing.md)")
        broken = check_links.find_broken(str(self.root))
        self.assertEqual(broken[0].source, "docs/a.md")


class SlugTest(unittest.TestCase):
    def test_matches_github_style_anchors(self) -> None:
        self.assertEqual(check_links.slug("6. Decisions"), "6-decisions")
        self.assertEqual(check_links.slug("Results and acceptance"), "results-and-acceptance")
        self.assertEqual(check_links.slug("The `veval` core (draft)"), "the-veval-core-draft")

    def test_duplicate_headings_get_numeric_suffixes_like_github(self) -> None:
        text = "# Notes\n\n## Step\n\n## Step\n\n## Step\n"
        self.assertEqual(check_links.heading_anchors(text), frozenset({"notes", "step", "step-1", "step-2"}))


class FindBrokenTest(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.root = Path(self.tmp.name)

    def tearDown(self) -> None:
        self.tmp.cleanup()

    def test_valid_relative_link_and_anchor_pass(self) -> None:
        write(self.root, "docs/a.md", "# Title\n\n## Results and acceptance\n")
        write(self.root, "README.md", "[design](docs/a.md#results-and-acceptance)")
        self.assertEqual(check_links.find_broken(str(self.root)), [])

    def test_missing_file_is_reported(self) -> None:
        write(self.root, "README.md", "[gone](docs/missing.md)")
        broken = check_links.find_broken(str(self.root))
        self.assertEqual(broken, [check_links.BrokenLink("README.md", "docs/missing.md", "missing file")])

    def test_missing_anchor_is_reported(self) -> None:
        write(self.root, "docs/a.md", "# Title\n\n## Old heading\n")
        write(self.root, "README.md", "[sec](docs/a.md#new-heading)")
        broken = check_links.find_broken(str(self.root))
        self.assertEqual(broken, [check_links.BrokenLink("README.md", "docs/a.md#new-heading", "missing anchor")])

    def test_anchor_into_a_non_markdown_target_is_reported_as_unchecked(self) -> None:
        write(self.root, "data.json", "{}")
        write(self.root, "README.md", "[frag](data.json#section)")
        broken = check_links.find_broken(str(self.root))
        self.assertEqual(broken, [check_links.BrokenLink("README.md", "data.json#section", "anchor not checkable")])

    def test_same_file_anchor_resolves(self) -> None:
        write(self.root, "a.md", "# Top\n\n## Details\n\nSee [details](#details).")
        self.assertEqual(check_links.find_broken(str(self.root)), [])

    def test_external_and_mailto_links_are_ignored(self) -> None:
        write(self.root, "a.md", "[w](https://example.org/nope) [m](mailto:x@y.z) [h](http://example.org)")
        self.assertEqual(check_links.find_broken(str(self.root)), [])

    def test_links_in_fenced_code_blocks_are_ignored(self) -> None:
        write(self.root, "a.md", "# A\n\n```text\n[fake](nope.md)\n```\n")
        self.assertEqual(check_links.find_broken(str(self.root)), [])

    def test_links_from_nested_documents_resolve_relative_to_their_directory(self) -> None:
        write(self.root, "docs/decisions/0001-x.md", "[req](../requirements.md#6-decisions)")
        write(self.root, "docs/requirements.md", "# Requirements\n\n## 6. Decisions\n")
        self.assertEqual(check_links.find_broken(str(self.root)), [])

    def test_skipped_directories_are_not_scanned(self) -> None:
        write(self.root, "node_modules/pkg/README.md", "[bad](nope.md)")
        write(self.root, ".git/notes.md", "[bad](nope.md)")
        write(self.root, "a.md", "# A\n")
        self.assertEqual(check_links.find_broken(str(self.root)), [])

    def test_missing_root_raises_instead_of_returning_nothing(self) -> None:
        with self.assertRaises(NotADirectoryError):
            check_links.find_broken(str(self.root / "missing"))


class StandaloneMainTest(unittest.TestCase):
    """check_links must be usable on its own, without vdocs or gen_sources."""

    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.root = Path(self.tmp.name)
        write(self.root, "a.md", "# A\n\n## Part\n\n[ok](b.md#b)")
        write(self.root, "b.md", "# B\n")

    def tearDown(self) -> None:
        self.tmp.cleanup()

    def run_main(self, argv: list[str]) -> tuple[int, str, str]:
        import contextlib
        import io

        out, err = io.StringIO(), io.StringIO()
        with contextlib.redirect_stdout(out), contextlib.redirect_stderr(err):
            code = check_links.main(argv)
        return code, out.getvalue(), err.getvalue()

    def test_each_function_has_its_own_subcommand(self) -> None:
        self.assertEqual(self.run_main(["slug", "6. Decisions"]), (0, "6-decisions\n", ""))
        self.assertEqual(self.run_main(["anchors", str(self.root / "a.md")]), (0, "a\npart\n", ""))
        code, out, _ = self.run_main(["check", str(self.root)])
        self.assertEqual((code, out), (0, "checked 2 files, 0 broken\n"))

    def test_check_exit_code_reflects_breakage(self) -> None:
        write(self.root, "a.md", "[bad](c.md)")
        code, out, _ = self.run_main(["check", str(self.root)])
        self.assertEqual(code, 1)
        self.assertIn("a.md: c.md (missing file)", out)

    def test_not_checkable_anchors_are_reported_but_do_not_fail_the_check(self) -> None:
        write(self.root, "data.json", "{}")
        write(self.root, "a.md", "# A\n\n[frag](data.json#s) [ok](b.md#b)")
        code, out, _ = self.run_main(["check", str(self.root)])
        self.assertEqual(code, 0)
        self.assertIn("a.md: data.json#s (anchor not checkable)", out)
        self.assertIn("1 not checkable", out)
        self.assertIn("checked 2 files, 0 broken", out)

    def test_errors_exit_2_with_a_message(self) -> None:
        code, _, err = self.run_main(["check", str(self.root / "missing")])
        self.assertEqual((code, "not a directory" in err), (2, True))
        with tempfile.TemporaryDirectory() as empty:
            code, _, err = self.run_main(["check", empty])
            self.assertEqual((code, "no Markdown" in err), (2, True))


if __name__ == "__main__":
    unittest.main()
