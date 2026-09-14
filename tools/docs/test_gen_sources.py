"""Tests for gen_sources: the source register must list every external link once,
grouped by domain, with every citing document, and must ignore its own output path."""
import os
import tempfile
import unittest
import unittest.mock
from pathlib import Path

import gen_sources


def write(root: Path, rel: str, text: str) -> None:
    path = root / rel
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(text, encoding="utf-8")


class CollectLinksTest(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.root = Path(self.tmp.name)

    def tearDown(self) -> None:
        self.tmp.cleanup()

    def test_titled_and_bare_links_are_collected(self) -> None:
        write(self.root, "a.md", "See [Paper](https://arxiv.org/abs/1) and https://example.org/x.")
        links = gen_sources.collect_links(str(self.root))
        self.assertEqual(links["https://arxiv.org/abs/1"].titles, frozenset({"Paper"}))
        self.assertEqual(links["https://example.org/x"].files, frozenset({"a.md"}))

    def test_same_url_in_two_files_is_listed_once_with_both_citers(self) -> None:
        write(self.root, "a.md", "[One](https://example.org/p)")
        write(self.root, "docs/b.md", "[Two](https://example.org/p)")
        links = gen_sources.collect_links(str(self.root))
        self.assertEqual(len(links), 1)
        self.assertEqual(links["https://example.org/p"].files, frozenset({"a.md", "docs/b.md"}))
        self.assertEqual(links["https://example.org/p"].titles, frozenset({"One", "Two"}))

    def test_bare_url_in_a_second_file_adds_that_file_as_a_citer(self) -> None:
        write(self.root, "a.md", "[Paper](https://example.org/p)")
        write(self.root, "b.md", "Also discussed at https://example.org/p today.")
        links = gen_sources.collect_links(str(self.root))
        self.assertEqual(links["https://example.org/p"].files, frozenset({"a.md", "b.md"}))
        self.assertEqual(links["https://example.org/p"].titles, frozenset({"Paper"}))

    def test_trailing_punctuation_is_stripped_from_bare_urls(self) -> None:
        write(self.root, "a.md", "Read https://example.org/q.")
        links = gen_sources.collect_links(str(self.root))
        self.assertIn("https://example.org/q", links)
        self.assertNotIn("https://example.org/q.", links)

    def test_register_output_and_skipped_directories_are_ignored(self) -> None:
        write(self.root, "docs/research/sources.md", "[Old](https://example.org/old)")
        write(self.root, "node_modules/x.md", "[Dep](https://example.org/dep)")
        write(self.root, ".github/y.md", "[Tpl](https://example.org/tpl)")
        write(self.root, "a.md", "[Keep](https://example.org/keep)")
        links = gen_sources.collect_links(str(self.root))
        self.assertEqual(set(links), {"https://example.org/keep", "https://example.org/tpl"})

    def test_superpowers_scratch_is_not_collected(self) -> None:
        write(self.root, ".superpowers/sdd/task-1-brief.md", "[Scratch](https://example.org/scratch)")
        write(self.root, "a.md", "[Keep](https://example.org/keep)")
        links = gen_sources.collect_links(str(self.root))
        self.assertEqual(set(links), {"https://example.org/keep"})

    def test_a_file_merely_named_sources_md_elsewhere_is_not_ignored(self) -> None:
        write(self.root, "notes/sources.md", "[Elsewhere](https://example.org/elsewhere)")
        links = gen_sources.collect_links(str(self.root))
        self.assertEqual(set(links), {"https://example.org/elsewhere"})

    def test_relative_links_are_not_external_sources(self) -> None:
        write(self.root, "a.md", "[Local](docs/b.md) and [Ext](https://example.org/e)")
        links = gen_sources.collect_links(str(self.root))
        self.assertEqual(set(links), {"https://example.org/e"})

    def test_missing_root_raises_instead_of_returning_nothing(self) -> None:
        with self.assertRaises(NotADirectoryError):
            gen_sources.markdown_files(str(self.root / "missing"))


class EdgeCaseTest(unittest.TestCase):
    """Cases the code review found silently mishandled."""

    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.root = Path(self.tmp.name)

    def tearDown(self) -> None:
        self.tmp.cleanup()

    def test_links_inside_code_fences_and_inline_code_are_not_citations(self) -> None:
        write(self.root, "a.md", "```md\n[fake](https://example.org/fenced)\n```\nInline `[x](https://example.org/inline)` and [Real](https://example.org/real)")
        self.assertEqual(set(gen_sources.collect_links(str(self.root))), {"https://example.org/real"})

    def test_autolinks_bracketed_targets_and_quoted_titles_are_collected(self) -> None:
        write(self.root, "a.md", "<https://example.org/auto> [B](<https://example.org/brack>) [Q](https://example.org/q \"Accessed 2026\")")
        links = gen_sources.collect_links(str(self.root))
        self.assertEqual(set(links), {"https://example.org/auto", "https://example.org/brack", "https://example.org/q"})
        self.assertEqual(links["https://example.org/q"].titles, frozenset({"Q"}))

    def test_unreadable_file_is_a_tool_error_with_the_path(self) -> None:
        (self.root / "bad.md").write_bytes(b"\xff\xfe not utf-8")
        with self.assertRaises(gen_sources.ToolError) as ctx:
            gen_sources.collect_links(str(self.root))
        self.assertIn("bad.md", str(ctx.exception))

    def test_dangling_symlink_is_a_tool_error_not_a_crash(self) -> None:
        try:
            os.symlink(str(self.root / "missing.md"), str(self.root / "dangling.md"))
        except (OSError, NotImplementedError):
            self.skipTest("symlinks not permitted here")
        with self.assertRaises(gen_sources.ToolError) as ctx:
            gen_sources.collect_links(str(self.root))
        self.assertIn("dangling.md", str(ctx.exception))

    def test_urls_with_balanced_parentheses_are_kept_whole(self) -> None:
        text = "[wiki](https://en.wikipedia.org/wiki/Bracket_(mathematics)) and bare https://en.wikipedia.org/wiki/Foo_(bar) too."
        self.assertEqual([u for u, _ in gen_sources.external_links(text)],
                         ["https://en.wikipedia.org/wiki/Bracket_(mathematics)", "https://en.wikipedia.org/wiki/Foo_(bar)"])

    def test_unterminated_and_longer_closing_fences_are_still_code(self) -> None:
        unterminated = "```md\n[fake](https://example.org/fenced)\nnever closed\n"
        self.assertEqual(gen_sources.external_links(unterminated), [])
        longer_close = "```\n[fake](https://example.org/fenced)\n`````\n[Real](https://example.org/real)"
        self.assertEqual(gen_sources.external_links(longer_close), [("https://example.org/real", "Real")])

    def test_cross_drive_output_only_excludes_the_canonical_register(self) -> None:
        with unittest.mock.patch.object(os.path, "relpath", side_effect=ValueError("different drives")):
            self.assertEqual(gen_sources.register_exclusions("C:/repo", "D:/out.md"), frozenset({gen_sources.OUTPUT_REL}))

    def test_root_level_file_whose_name_starts_with_dots_is_still_inside_the_root(self) -> None:
        self.assertIn("..notes.md", gen_sources.register_exclusions(str(self.root), str(self.root / "..notes.md")))
        self.assertNotIn("../outside.md", gen_sources.register_exclusions(str(self.root), str(self.root.parent / "outside.md")))

    @unittest.expectedFailure
    def test_fence_indented_inside_a_list_item_is_still_code(self) -> None:
        """Known limitation: fences indented four or more columns inside list items are not recognized."""
        self.assertEqual(gen_sources.external_links("- item\n\n    ```\n    https://example.org/code\n    ```\n"), [])

    @unittest.expectedFailure
    def test_url_with_a_literal_quote_is_kept_whole(self) -> None:
        """Known limitation: a raw apostrophe ends the URL; wrap such URLs in angle brackets."""
        self.assertEqual(gen_sources.external_links("https://en.wikipedia.org/wiki/Alzheimer's_disease"),
                         [("https://en.wikipedia.org/wiki/Alzheimer's_disease", "")])


class DomainTest(unittest.TestCase):
    def test_host_is_lowercased(self) -> None:
        self.assertEqual(gen_sources.domain("https://Example.ORG/A"), "example.org")

    def test_scheme_and_www_are_dropped(self) -> None:
        self.assertEqual(gen_sources.domain("https://www.example.org/a/b"), "example.org")
        self.assertEqual(gen_sources.domain("http://arxiv.org/abs/1"), "arxiv.org")


class RenderRegisterTest(unittest.TestCase):
    def test_groups_by_domain_in_sorted_order_and_labels_by_title(self) -> None:
        links = {
            "https://zeta.org/1": gen_sources.Citation(titles=frozenset({"Zeta"}), files=frozenset({"a.md"})),
            "https://alpha.org/2": gen_sources.Citation(titles=frozenset(), files=frozenset({"b.md", "a.md"})),
        }
        text = gen_sources.render_register(links, file_count=2)
        self.assertLess(text.index("## alpha.org"), text.index("## zeta.org"))
        self.assertIn("- [Zeta](https://zeta.org/1) — cited in `a.md`", text)
        self.assertIn("- [https://alpha.org/2](https://alpha.org/2) — cited in `a.md`, `b.md`", text)
        self.assertIn("Unique external links: 2. Citing documents: 2.", text)

    def test_register_text_is_deterministic_for_a_root(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            write(root, "a.md", "[Paper](https://arxiv.org/abs/1)")
            self.assertEqual(gen_sources.register_text(str(root)), gen_sources.register_text(str(root)))
            self.assertTrue(gen_sources.register_text(str(root)).startswith("# Source register"))


class StandaloneMainTest(unittest.TestCase):
    """gen_sources must be usable on its own, without vdocs or check_links."""

    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.root = Path(self.tmp.name)
        write(self.root, "a.md", "[Paper](https://arxiv.org/abs/1)")
        (self.root / "docs" / "research").mkdir(parents=True)

    def tearDown(self) -> None:
        self.tmp.cleanup()

    def run_main(self, argv: list[str]) -> tuple[int, str, str]:
        import contextlib
        import io

        out, err = io.StringIO(), io.StringIO()
        with contextlib.redirect_stdout(out), contextlib.redirect_stderr(err):
            code = gen_sources.main(argv)
        return code, out.getvalue(), err.getvalue()

    def test_each_function_has_its_own_subcommand(self) -> None:
        self.assertEqual(self.run_main(["files", str(self.root)]), (0, "a.md\n", ""))
        self.assertEqual(self.run_main(["links", str(self.root / "a.md")]), (0, "https://arxiv.org/abs/1\tPaper\n", ""))
        self.assertEqual(self.run_main(["domain", "https://www.arxiv.org/abs/1"]), (0, "arxiv.org\n", ""))
        code, out, _ = self.run_main(["register", str(self.root)])
        self.assertEqual(code, 0)
        self.assertIn("wrote", out)
        self.assertEqual(self.run_main(["register", str(self.root), "--check"])[0], 0)

    def test_register_check_reports_missing_and_stale(self) -> None:
        code, _, err = self.run_main(["register", str(self.root), "--check"])
        self.assertEqual((code, "missing" in err), (1, True))
        self.run_main(["register", str(self.root)])
        write(self.root, "b.md", "[New](https://example.org/new)")
        code, _, err = self.run_main(["register", str(self.root), "--check"])
        self.assertEqual((code, "stale" in err), (1, True))

    def test_custom_output_inside_root_never_cites_itself_and_converges(self) -> None:
        output = str(self.root / "docs" / "alt-register.md")
        self.assertEqual(self.run_main(["register", str(self.root), "--output", output])[0], 0)
        self.assertEqual(self.run_main(["register", str(self.root), "--output", output, "--check"])[0], 0)
        self.assertNotIn("alt-register.md`", (self.root / "docs" / "alt-register.md").read_text(encoding="utf-8"))

    def test_errors_exit_2_with_a_message(self) -> None:
        code, _, err = self.run_main(["files", str(self.root / "missing")])
        self.assertEqual((code, "not a directory" in err), (2, True))
        with tempfile.TemporaryDirectory() as empty:
            code, _, err = self.run_main(["register", empty])
            self.assertEqual((code, "no Markdown" in err), (2, True))


if __name__ == "__main__":
    unittest.main()
