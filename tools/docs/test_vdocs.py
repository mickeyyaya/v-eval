"""Tests for the vdocs CLI: every capability of the two tools is reachable as a subcommand,
errors fail loudly with exit code 2, and the register check detects drift."""
import contextlib
import io
import os
import tempfile
import unittest
from pathlib import Path

import vdocs


def write(root: Path, rel: str, text: str) -> None:
    path = root / rel
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(text, encoding="utf-8")


def run(argv: list[str]) -> tuple[int, str, str]:
    out, err = io.StringIO(), io.StringIO()
    with contextlib.redirect_stdout(out), contextlib.redirect_stderr(err):
        code = vdocs.main(argv)
    return code, out.getvalue(), err.getvalue()


class RepoFixture(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.root = Path(self.tmp.name)
        write(self.root, "README.md", "# Top\n\n## Usage\n\n[design](docs/a.md#results) and [Paper](https://arxiv.org/abs/1)")
        write(self.root, "docs/a.md", "# A\n\n## Results\n\nSee https://example.org/x.")
        (self.root / "docs" / "research").mkdir(parents=True)

    def tearDown(self) -> None:
        self.tmp.cleanup()


class FilesAndLinksTest(RepoFixture):
    def test_files_lists_markdown_documents_in_sorted_order(self) -> None:
        code, out, _ = run(["files", str(self.root)])
        self.assertEqual(code, 0)
        self.assertEqual(out.splitlines(), ["README.md", "docs/a.md"])

    def test_links_prints_url_and_title_per_line(self) -> None:
        code, out, _ = run(["links", str(self.root / "README.md")])
        self.assertEqual(code, 0)
        self.assertEqual(out.splitlines(), ["https://arxiv.org/abs/1\tPaper"])

    def test_domain_and_slug_are_exposed(self) -> None:
        self.assertEqual(run(["domain", "https://www.example.org/a"]), (0, "example.org\n", ""))
        self.assertEqual(run(["slug", "6. Decisions"]), (0, "6-decisions\n", ""))

    def test_anchors_lists_heading_slugs(self) -> None:
        code, out, _ = run(["anchors", str(self.root / "docs" / "a.md")])
        self.assertEqual(code, 0)
        self.assertEqual(out.splitlines(), ["a", "results"])


class RegisterTest(RepoFixture):
    def test_register_writes_file_and_check_passes_when_current(self) -> None:
        code, out, _ = run(["register", str(self.root)])
        self.assertEqual(code, 0)
        self.assertIn("wrote", out)
        register = (self.root / "docs" / "research" / "sources.md").read_text(encoding="utf-8")
        self.assertIn("https://arxiv.org/abs/1", register)
        self.assertIn("https://example.org/x", register)
        self.assertEqual(run(["register", str(self.root), "--check"])[0], 0)

    def test_register_check_fails_when_stale_or_missing(self) -> None:
        code, _, err = run(["register", str(self.root), "--check"])
        self.assertEqual(code, 1)
        self.assertIn("missing", err)
        run(["register", str(self.root)])
        write(self.root, "docs/b.md", "[New](https://example.org/new)")
        code, _, err = run(["register", str(self.root), "--check"])
        self.assertEqual(code, 1)
        self.assertIn("stale", err)

    def test_register_excludes_only_its_own_output_path(self) -> None:
        write(self.root, "docs/other/sources.md", "[Other](https://example.org/other)")
        run(["register", str(self.root)])
        register = (self.root / "docs" / "research" / "sources.md").read_text(encoding="utf-8")
        self.assertIn("https://example.org/other", register)
        self.assertNotIn("docs/research/sources.md`", register.replace(os.sep, "/"))


class RegisterDirectoryTest(unittest.TestCase):
    def test_register_creates_the_output_directory(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write(root, "a.md", "[P](https://example.org/p)")
            self.assertEqual(run(["register", str(root)])[0], 0)
            self.assertTrue((root / "docs" / "research" / "sources.md").exists())


class CheckLinksTest(RepoFixture):
    def test_check_links_passes_and_fails_with_exit_codes(self) -> None:
        self.assertEqual(run(["check-links", str(self.root)])[0], 0)
        write(self.root, "docs/a.md", "# A\n\n## Results\n\n[bad](nope.md)")
        code, out, _ = run(["check-links", str(self.root)])
        self.assertEqual(code, 1)
        self.assertIn("nope.md (missing file)", out)
        self.assertIn("checked 2 files, 1 broken", out)


class FailLoudlyTest(unittest.TestCase):
    def test_missing_root_is_an_error_not_a_pass(self) -> None:
        code, _, err = run(["check-links", "/definitely/not/a/dir"])
        self.assertEqual(code, 2)
        self.assertIn("not a directory", err)

    def test_root_without_markdown_is_an_error_not_a_pass(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            for argv in (["check-links", tmp], ["register", tmp], ["files", tmp]):
                code, _, err = run(argv)
                self.assertEqual(code, 2, argv)
                self.assertIn("no Markdown", err)

    def test_unknown_subcommand_and_stray_flags_are_rejected(self) -> None:
        with self.assertRaises(SystemExit) as ctx:
            run(["frobnicate"])
        self.assertEqual(ctx.exception.code, 2)
        with self.assertRaises(SystemExit):
            run(["check-links", "--bogus"])


if __name__ == "__main__":
    unittest.main()
