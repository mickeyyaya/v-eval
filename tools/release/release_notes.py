#!/usr/bin/env python3
"""Release notes for one version, read from a Keep a Changelog file.

Prints the body of the ``## [X.Y.Z] - <date>`` section (``## [X.Y.Z]`` and ``## X.Y.Z`` are
accepted too) up to the next level-two heading or the link-reference block at the end of the
file, trimmed, with one trailing newline. Headings and link references inside fenced code
blocks are not boundaries. ``-o FILE`` writes the notes to a file with LF line endings instead
of standard output, creating the parent directory. Exits 2 with a message on standard error
when the version has no section, the section is empty, or the file cannot be read, so an
undocumented or empty tag fails the release workflow (decision 0025). A link whose target is
relative -- a path in the repository rather than a URL, a mailto, or an anchor -- resolves on
the changelog's own page but not on the release page, where the notes stand alone, so each one
is reported as a warning on standard error; the notes and the exit code are unchanged.

Repository maintenance tool only (decision 0022): standard library only; runs on macOS,
Linux, and Windows with ``python tools/release/release_notes.py CHANGELOG.md 0.1.0``.
"""
from __future__ import annotations

import argparse
import os
import re
import sys
from typing import Sequence

FENCE_OPEN = re.compile(r"^[ \t]{0,3}(`{3,}|~{3,})")  # kept identical to tools/docs/gen_sources.py and tools/docs/check_links.py on purpose
LEVEL_TWO = re.compile(r"^##[ \t]")
LINK_REFERENCE = re.compile(r"^\[[^\]]+\]:[ \t]+\S")
INLINE_LINK = re.compile(r"\[[^\]]*\]\(([^)\s]+)(?:[ \t]+\"[^\"]*\")?\)")  # [text](target "title"), images included
ABSOLUTE_PREFIXES = ("http://", "https://", "mailto:", "#")


class ToolError(Exception):
    """A user-facing failure: reported on stderr with exit code 2."""


def read_text(path: str) -> str:
    try:
        with open(path, encoding="utf-8") as handle:
            return handle.read()
    except UnicodeDecodeError as exc:
        raise ToolError(f"cannot read {path}: not valid UTF-8 ({exc.reason} at byte {exc.start})") from exc
    except OSError as exc:
        raise ToolError(f"cannot read {path}: {exc.strerror}") from exc


def prose_lines(text: str) -> list[tuple[str, bool]]:
    """Every line paired with whether it is prose (True) or inside a fenced code block (False).
    CommonMark fences: the closer uses the same character and is at least as long as the opener;
    an unterminated fence runs to the end of the document."""
    lines: list[tuple[str, bool]] = []
    opener: str | None = None
    for line in text.split("\n"):
        match = FENCE_OPEN.match(line)
        fence = match.group(1) if match else None
        if opener is None:
            lines.append((line, not fence))
            opener = fence
        else:
            lines.append((line, False))
            # fence is a run of one repeated character, so a matching prefix implies same character and sufficient length
            if fence and line.strip() == fence and fence[: len(opener)] == opener:
                opener = None
    return lines


def heading_pattern(version: str) -> re.Pattern[str]:
    """The level-two heading that opens the section: the version token, bracketed or bare,
    optionally followed by whitespace and anything (the date, a yanked mark)."""
    token = re.escape(version)
    return re.compile(rf"^##[ \t]+(?:\[{token}\]|{token})(?:[ \t].*)?$")


def section(text: str, version: str) -> str | None:
    """The trimmed body of the version's section with one trailing newline, or None when the
    file has no such section."""
    heading = heading_pattern(version)
    lines = prose_lines(text)
    start = next((i for i, (line, prose) in enumerate(lines) if prose and heading.match(line)), None)
    if start is None:
        return None
    body: list[str] = []
    for line, prose in lines[start + 1 :]:
        if prose and (LEVEL_TWO.match(line) or LINK_REFERENCE.match(line)):
            break
        body.append(line)
    return "\n".join(body).strip() + "\n"


def relative_links(notes: str) -> list[str]:
    """The target of every inline link in the notes' prose, in order, that is neither a URL,
    a mailto, nor an anchor: a repository path the release page has no way to resolve."""
    targets: list[str] = []
    for line, prose in prose_lines(notes):
        if not prose:
            continue
        for match in INLINE_LINK.finditer(line):
            target = match.group(1)
            if not target.lower().startswith(ABSOLUTE_PREFIXES):
                targets.append(target)
    return targets


def warn_relative_links(notes: str) -> None:
    for target in relative_links(notes):
        print(f'warning: relative link "{target}" will not resolve on the release page', file=sys.stderr)


def write_notes(path: str, notes: str) -> None:
    parent = os.path.dirname(path)
    if parent:
        os.makedirs(parent, exist_ok=True)
    with open(path, "w", encoding="utf-8", newline="\n") as handle:
        handle.write(notes)


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="release_notes", description=__doc__.split("\n\n")[0])
    parser.add_argument("changelog", help="path to the Keep a Changelog file")
    parser.add_argument("version", help="the version whose section to print, as it appears in the heading")
    parser.add_argument("-o", "--output", metavar="FILE", help="write the notes to FILE (LF line endings) instead of standard output")
    return parser


def main(argv: Sequence[str] | None = None) -> int:
    args = build_parser().parse_args(argv)
    try:
        notes = section(read_text(args.changelog), args.version)
        if notes is None:
            raise ToolError(f"no section for version {args.version} in {args.changelog}")
        if not notes.strip():
            raise ToolError(f"section for version {args.version} in {args.changelog} is empty")
        warn_relative_links(notes)
        if args.output:
            write_notes(args.output, notes)
        else:
            sys.stdout.write(notes)
    except ToolError as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 2
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
