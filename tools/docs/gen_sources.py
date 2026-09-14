#!/usr/bin/env python3
"""Source register for the documentation: every external link cited in the repository's
Markdown, deduplicated, grouped by domain, with the documents that cite it.

Library and standalone command line. Subcommands: ``files``, ``links``, ``domain``,
``register`` (with ``--check`` to verify the committed register is current). Output paths are
always written with forward slashes so the register is byte-identical on every platform.

Repository maintenance tool only (decision 0022): not part of the v-eval product, never on
the skill's required path. Standard library only; runs on macOS, Linux, and Windows with
``python tools/docs/gen_sources.py``.
"""
from __future__ import annotations

import argparse
import os
import re
import sys
from collections import defaultdict
from dataclasses import dataclass
from typing import Callable, Mapping, Sequence

URL_BODY = r"(?:[^\s<>()\"']|\([^\s<>()]*\))+"  # one level of balanced parentheses, as in Wikipedia URLs;
# a literal quote ends a URL (it may start a Markdown title), so wrap such URLs in angle brackets
TITLED_LINK = re.compile(r"\[([^\]]*)\]\(\s*<?(https?://" + URL_BODY + r")>?(?:\s+[\"'][^\"']*[\"'])?\s*\)")
AUTOLINK = re.compile(r"<(https?://[^>\s]+)>")
BARE_URL_BODY = r"(?:[^\s<>()\]\"']|\([^\s<>()\]]*\))+"  # as URL_BODY, but a bare URL also stops at a closing bracket
BARE_URL = re.compile(r"(?<![(\[<\"'])(https?://" + BARE_URL_BODY + r")")
INLINE_CODE = re.compile(r"`[^`\n]*`")
SKIP_DIRS = frozenset({".git", "node_modules", ".superpowers", "build"})  # kept identical to check_links.py on purpose
OUTPUT_REL = "docs/research/sources.md"
TRAILING_PUNCTUATION = ".,;"
HEADER = (
    "# Source register",
    "",
    "Generated from every Markdown file in the repository. Each external link cited anywhere "
    "in the documentation appears once, with the documents that cite it. All links were accessed "
    "on the research dates stated in the citing documents; this register does not re-verify them. "
    "Regenerate with `python tools/docs/vdocs.py register .` from the repository root; "
    "`--check` verifies the committed file is current.",
    "",
)


class ToolError(Exception):
    """A user-facing failure: reported on stderr with exit code 2."""


@dataclass(frozen=True)
class Citation:
    """Where one external URL is cited: the link titles used and the citing documents."""

    titles: frozenset[str] = frozenset()
    files: frozenset[str] = frozenset()

    def cite(self, file: str, title: str) -> Citation:
        titles = self.titles | {title} if title else self.titles
        return Citation(titles=titles, files=self.files | {file})


def to_posix(path: str) -> str:
    return path.replace(os.sep, "/")


def markdown_files(root: str, skip: frozenset[str] = SKIP_DIRS, exclude: frozenset[str] = frozenset({OUTPUT_REL})) -> list[str]:
    """Repository-relative, forward-slash paths of Markdown files, excluding skipped directories and the register(s)."""
    if not os.path.isdir(root):
        raise NotADirectoryError(f"{root} is not a directory")
    found: list[str] = []
    for dirpath, dirnames, filenames in os.walk(root):
        dirnames[:] = [d for d in dirnames if d not in skip]
        for name in filenames:
            rel = to_posix(os.path.relpath(os.path.join(dirpath, name), root))
            if name.endswith(".md") and rel not in exclude:
                found.append(rel)
    return sorted(found)


def require_markdown_files(root: str, exclude: frozenset[str] = frozenset({OUTPUT_REL})) -> list[str]:
    files = markdown_files(root, exclude=exclude)
    if not files:
        raise ToolError(f"no Markdown files under {root}")
    return files


def read_text(path: str) -> str:
    try:
        with open(path, encoding="utf-8") as handle:
            return handle.read()
    except UnicodeDecodeError as exc:
        raise ToolError(f"cannot read {path}: not valid UTF-8 ({exc.reason} at byte {exc.start})") from exc
    except OSError as exc:
        raise ToolError(f"cannot read {path}: {exc.strerror}") from exc


FENCE_OPEN = re.compile(r"^[ \t]{0,3}(`{3,}|~{3,})")  # kept identical to check_links.py on purpose


def strip_fenced(text: str) -> str:
    """Drop fenced code blocks, CommonMark style: the closer uses the same character and is at
    least as long as the opener; an unterminated fence runs to the end of the document.
    Known limitation: fences indented four or more columns inside list items are not recognized."""
    kept: list[str] = []
    opener: str | None = None
    for line in text.split("\n"):
        match = FENCE_OPEN.match(line)
        fence = match.group(1) if match else None
        if opener is None:
            if fence:
                opener = fence
            else:
                kept.append(line)
        # fence is a run of one repeated character, so a matching prefix implies same character and sufficient length
        elif fence and line.strip() == fence and fence[: len(opener)] == opener:
            opener = None
    return "\n".join(kept)


def strip_code(text: str) -> str:
    """Remove fenced blocks and inline code spans, where links are literal text, not citations."""
    return INLINE_CODE.sub("", strip_fenced(text))


def external_links(text: str) -> list[tuple[str, str]]:
    """Ordered unique (url, title) pairs in one document; autolinks and bare URLs carry an empty title."""
    prose = strip_code(text)
    titled = [(url, title.strip()) for title, url in TITLED_LINK.findall(prose)]
    untitled = [(url, "") for url in AUTOLINK.findall(prose) + BARE_URL.findall(prose)]
    seen: set[tuple[str, str]] = set()
    ordered: list[tuple[str, str]] = []
    for url, title in titled + untitled:
        pair = (url.rstrip(TRAILING_PUNCTUATION), title)
        if pair not in seen:
            seen.add(pair)
            ordered.append(pair)
    return ordered


def collect_links(root: str, exclude: frozenset[str] = frozenset({OUTPUT_REL})) -> dict[str, Citation]:
    """Every external URL across the repository, with all of its citers and titles."""
    by_url: dict[str, Citation] = {}
    for rel in markdown_files(root, exclude=exclude):
        for url, title in external_links(read_text(os.path.join(root, rel))):
            by_url[url] = by_url.get(url, Citation()).cite(rel, title)
    return by_url


def domain(url: str) -> str:
    return re.sub(r"^https?://(www\.)?", "", url).split("/")[0].lower()


def render_entry(url: str, citation: Citation) -> str:
    titles = sorted(t for t in citation.titles if t)
    label = titles[0] if titles else url
    cited = ", ".join(f"`{f}`" for f in sorted(citation.files))
    return f"- [{label}]({url}) — cited in {cited}"


def render_register(links: Mapping[str, Citation], file_count: int) -> str:
    """The register text: header, counts, then one section per domain in sorted order."""
    by_domain: dict[str, list[str]] = defaultdict(list)
    for url in links:
        by_domain[domain(url)].append(url)
    lines = [*HEADER, f"Unique external links: {len(links)}. Citing documents: {file_count}.", ""]
    for dom in sorted(by_domain):
        lines += [f"## {dom}", ""]
        lines += [render_entry(url, links[url]) for url in sorted(by_domain[dom])]
        lines.append("")
    return "\n".join(lines)


def register_text(root: str) -> str:
    files = require_markdown_files(root)
    return render_register(collect_links(root), len(files))


def register_status(output: str, expected: str) -> str:
    """'missing', 'stale', or 'current' for the register file on disk."""
    if not os.path.exists(output):
        return "missing"
    return "current" if read_text(output) == expected else "stale"


def cmd_files(args: argparse.Namespace) -> int:
    for rel in require_markdown_files(args.root):
        print(rel)
    return 0


def cmd_links(args: argparse.Namespace) -> int:
    for url, title in external_links(read_text(args.file)):
        print(f"{url}\t{title}")
    return 0


def cmd_domain(args: argparse.Namespace) -> int:
    print(domain(args.url))
    return 0


def register_exclusions(root: str, output: str) -> frozenset[str]:
    """The canonical register plus the chosen output when it lies inside the root, so a register never cites itself."""
    try:
        rel = to_posix(os.path.relpath(output, root))
    except ValueError:  # Windows: output on another drive, so it cannot lie inside the root
        return frozenset({OUTPUT_REL})
    inside = rel != ".." and not rel.startswith("../")
    return frozenset({OUTPUT_REL}) | (frozenset({rel}) if inside else frozenset())


def cmd_register(args: argparse.Namespace) -> int:
    output = args.output or os.path.join(args.root, *OUTPUT_REL.split("/"))
    exclude = register_exclusions(args.root, output)
    files = require_markdown_files(args.root, exclude)
    links = collect_links(args.root, exclude)
    expected = render_register(links, len(files))
    if args.check:
        status = register_status(output, expected)
        if status != "current":
            print(f"register {status}: {output}", file=sys.stderr)
            return 1
        print(f"register current: {output}")
        return 0
    os.makedirs(os.path.dirname(output) or ".", exist_ok=True)
    with open(output, "w", encoding="utf-8", newline="\n") as handle:
        handle.write(expected)
    print(f"wrote {output}: {len(links)} unique links from {len(files)} files")
    return 0


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="gen_sources", description=__doc__.split("\n\n")[0])
    sub = parser.add_subparsers(dest="command", required=True)
    files = sub.add_parser("files", help="list the Markdown files the register covers")
    files.add_argument("root", nargs="?", default=".")
    files.set_defaults(run=cmd_files)
    links = sub.add_parser("links", help="list external links in one document as url<TAB>title")
    links.add_argument("file")
    links.set_defaults(run=cmd_links)
    dom = sub.add_parser("domain", help="print the domain used to group a URL")
    dom.add_argument("url")
    dom.set_defaults(run=cmd_domain)
    reg = sub.add_parser("register", help="write the source register, or verify it with --check")
    reg.add_argument("root", nargs="?", default=".")
    reg.add_argument("--check", action="store_true", help="exit 1 if the register is missing or stale")
    reg.add_argument("--output", help=f"register path (default: <root>/{OUTPUT_REL})")
    reg.set_defaults(run=cmd_register)
    return parser


def main(argv: Sequence[str] | None = None) -> int:
    args = build_parser().parse_args(argv)
    run: Callable[[argparse.Namespace], int] = args.run
    try:
        return run(args)
    except (NotADirectoryError, ToolError) as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 2


if __name__ == "__main__":
    raise SystemExit(main())
