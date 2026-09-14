#!/usr/bin/env python3
"""Local check that relative Markdown links and heading anchors resolve.

Library and standalone command line. Subcommands: ``slug``, ``anchors``, ``check``.
``check`` prints one line per finding and a summary; anchors it cannot check (fragments into
non-Markdown targets) are listed as not checkable and never counted as broken; exits 1 only
when something is broken.

Advisory tool for a fast pre-commit pass. The authority in continuous integration is lychee
(``.github/workflows/docs.yml``); this checker uses a simplified GitHub-style anchor algorithm
(lowercase, punctuation dropped, spaces to hyphens, ``-N`` suffixes for repeated headings) and
may disagree with lychee on unusual headings; reference-style links (``[text][ref]``) are not
checked. Repository maintenance tool only (decision 0022);
standard library only; runs on macOS, Linux, and Windows with ``python tools/docs/check_links.py``.
"""
from __future__ import annotations

import argparse
import os
import posixpath
import re
import sys
from dataclasses import dataclass
from typing import Callable, Mapping, Sequence
from urllib.parse import unquote

SKIP_DIRS = frozenset({".git", "node_modules", ".superpowers", "build"})  # kept identical to gen_sources.py on purpose
LINK_TARGET = re.compile(r"\[[^\]]*\]\(\s*(?:<([^>]+)>|((?:[^\s()]|\([^\s()]*\))+))")  # group 1: <...> target; group 2: bare target, one level of balanced parentheses
HEADING = re.compile(r"^#{1,6}\s+(.+?)\s*#*\s*$", re.M)
INLINE_CODE = re.compile(r"`[^`\n]*`")
EXTERNAL_PREFIXES = ("http://", "https://", "mailto:")
NOT_CHECKABLE = "anchor not checkable"  # an unknown, reported separately, never counted as broken
ABSOLUTE_TARGET = "absolute path not checkable"  # leading "/" has no repository meaning; also an unknown
UNKNOWN_REASONS = frozenset({NOT_CHECKABLE, ABSOLUTE_TARGET})


class ToolError(Exception):
    """A user-facing failure: reported on stderr with exit code 2."""


@dataclass(frozen=True)
class BrokenLink:
    source: str
    target: str
    reason: str


def slug(heading: str) -> str:
    """GitHub-style anchor for one heading (without the duplicate suffix)."""
    text = re.sub(r"[`*_]", "", heading.strip().lower())
    text = re.sub(r"[^\w\- ]", "", text)
    return re.sub(r"\s+", "-", text)


def to_posix(path: str) -> str:
    return path.replace(os.sep, "/")


def markdown_files(root: str, skip: frozenset[str] = SKIP_DIRS) -> list[str]:
    """Repository-relative, forward-slash paths of Markdown files, excluding skipped directories."""
    if not os.path.isdir(root):
        raise NotADirectoryError(f"{root} is not a directory")
    found: list[str] = []
    for dirpath, dirnames, filenames in os.walk(root):
        dirnames[:] = [d for d in dirnames if d not in skip]
        found += [to_posix(os.path.relpath(os.path.join(dirpath, n), root)) for n in filenames if n.endswith(".md")]
    return sorted(found)


def read_text(path: str) -> str:
    try:
        with open(path, encoding="utf-8") as handle:
            return handle.read()
    except UnicodeDecodeError as exc:
        raise ToolError(f"cannot read {path}: not valid UTF-8 ({exc.reason} at byte {exc.start})") from exc
    except OSError as exc:
        raise ToolError(f"cannot read {path}: {exc.strerror}") from exc


FENCE_OPEN = re.compile(r"^[ \t]{0,3}(`{3,}|~{3,})")  # kept identical to gen_sources.py on purpose


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


def disk_paths(root: str) -> frozenset[str]:
    """Every file and directory under the root, as forward-slash relative paths spelled exactly as
    on disk. One walk serves every link, and membership is case-exact on every filesystem; a link
    that escapes the root is never a member."""
    paths: set[str] = set()
    for dirpath, dirnames, filenames in os.walk(root):
        dirnames[:] = [d for d in dirnames if d not in SKIP_DIRS]
        for name in dirnames + filenames:
            paths.add(to_posix(os.path.relpath(os.path.join(dirpath, name), root)))
    return frozenset(paths)


def heading_anchors(text: str) -> frozenset[str]:
    """Anchors of all headings outside code fences, with GitHub's -1, -2 suffixes for repeats."""
    counts: dict[str, int] = {}
    anchors: list[str] = []
    for heading in HEADING.findall(strip_fenced(text)):
        base = slug(heading)
        seen = counts.get(base, 0)
        counts[base] = seen + 1
        anchors.append(base if seen == 0 else f"{base}-{seen}")
    return frozenset(anchors)


def link_targets(text: str) -> list[str]:
    """Link targets in prose, with fenced blocks and inline code removed."""
    prose = INLINE_CODE.sub("", strip_fenced(text))
    return [bracketed or plain for bracketed, plain in LINK_TARGET.findall(prose)]


def resolve(source: str, target: str) -> tuple[str, str]:
    """(destination document in forward-slash form, fragment) for a relative link target."""
    path, _, fragment = target.partition("#")
    if not path:
        return source, fragment
    dest = posixpath.normpath(posixpath.join(posixpath.dirname(source), unquote(path)))
    return dest, fragment


def classify(source: str, target: str, on_disk: frozenset[str], anchors: Mapping[str, frozenset[str]]) -> str | None:
    """The reason a link is broken or unknown, or None when it resolves."""
    if target.startswith("/"):
        return ABSOLUTE_TARGET
    dest, fragment = resolve(source, target)
    if dest != source and dest.rstrip("/") not in on_disk:
        return "missing file"
    if not fragment:
        return None
    if dest not in anchors:
        return NOT_CHECKABLE
    return None if fragment in anchors[dest] else "missing anchor"


def check_document(source: str, text: str, on_disk: frozenset[str], anchors: Mapping[str, frozenset[str]]) -> list[BrokenLink]:
    """Broken or unknown relative links in one document; external links are ignored."""
    broken: list[BrokenLink] = []
    for target in link_targets(text):
        if target.startswith(EXTERNAL_PREFIXES):
            continue
        reason = classify(source, target, on_disk, anchors)
        if reason:
            broken.append(BrokenLink(source, target, reason))
    return broken


def find_broken(root: str) -> list[BrokenLink]:
    files = markdown_files(root)
    texts = {rel: read_text(os.path.join(root, rel)) for rel in files}
    anchors = {rel: heading_anchors(text) for rel, text in texts.items()}
    on_disk = disk_paths(root)
    return [link for rel in files for link in check_document(rel, texts[rel], on_disk, anchors)]


def cmd_slug(args: argparse.Namespace) -> int:
    print(slug(args.heading))
    return 0


def cmd_anchors(args: argparse.Namespace) -> int:
    for anchor in sorted(heading_anchors(read_text(args.file))):
        print(anchor)
    return 0


def cmd_check(args: argparse.Namespace) -> int:
    files = markdown_files(args.root)
    if not files:
        raise ToolError(f"no Markdown files under {args.root}")
    findings = find_broken(args.root)
    for link in findings:
        print(f"{link.source}: {link.target} ({link.reason})")
    broken = [link for link in findings if link.reason not in UNKNOWN_REASONS]
    if len(findings) != len(broken):
        print(f"{len(findings) - len(broken)} not checkable")
    print(f"checked {len(files)} files, {len(broken)} broken")
    return 1 if broken else 0


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="check_links", description=__doc__.split("\n\n")[0])
    sub = parser.add_subparsers(dest="command", required=True)
    slug_cmd = sub.add_parser("slug", help="print the anchor for a heading")
    slug_cmd.add_argument("heading")
    slug_cmd.set_defaults(run=cmd_slug)
    anchors = sub.add_parser("anchors", help="list the anchors defined by one document")
    anchors.add_argument("file")
    anchors.set_defaults(run=cmd_anchors)
    check = sub.add_parser("check", help="report broken relative links and anchors under a root")
    check.add_argument("root", nargs="?", default=".")
    check.set_defaults(run=cmd_check)
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
