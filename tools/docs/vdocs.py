#!/usr/bin/env python3
"""One front door for the documentation maintenance tools.

Every capability of ``gen_sources`` and ``check_links`` is reachable here as a subcommand, and
each of those modules also runs on its own; this file only routes. Subcommands: ``files``,
``links``, ``domain``, ``register`` (``--check``, ``--output``), ``slug``, ``anchors``,
``check-links``. Exit codes: 0 ok, 1 a check failed, 2 the tool could not run.

Repository maintenance tool only (decision 0022); standard library only; runs on macOS, Linux,
and Windows with ``python tools/docs/vdocs.py``.
"""
from __future__ import annotations

import argparse
import sys
from typing import Callable, Sequence

import check_links
import gen_sources

# Non-register commands: vdocs subcommand -> (owning module's main, its subcommand name,
# the vdocs argument that carries the one positional value that subcommand takes).
_FORWARDS: dict[str, tuple[Callable[[list[str]], int], str, str]] = {
    "files": (gen_sources.main, "files", "root"),
    "links": (gen_sources.main, "links", "file"),
    "domain": (gen_sources.main, "domain", "url"),
    "check-links": (check_links.main, "check", "root"),
    "slug": (check_links.main, "slug", "heading"),
    "anchors": (check_links.main, "anchors", "file"),
}


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="vdocs", description=__doc__.split("\n\n")[0])
    sub = parser.add_subparsers(dest="command", required=True)
    for name, help_text in (("files", "list the Markdown files the register covers (the register itself excluded)"),
                            ("register", "write the source register, or verify it with --check"),
                            ("check-links", "check relative links and anchors between documents (not external URLs)")):
        cmd = sub.add_parser(name, help=help_text)
        cmd.add_argument("root", nargs="?", default=".")
        if name == "register":
            cmd.add_argument("--check", action="store_true")
            cmd.add_argument("--output")
    sub.add_parser("links", help="list external URLs cited in one document, as url<TAB>title").add_argument("file")
    sub.add_parser("domain", help="print the domain used to group a URL").add_argument("url")
    sub.add_parser("slug", help="print the anchor for a heading").add_argument("heading")
    sub.add_parser("anchors", help="list the anchors defined by one document").add_argument("file")
    return parser


def route(args: argparse.Namespace) -> int:
    """Translate a vdocs invocation into the owning module's own command line."""
    if args.command == "register":
        argv = ["register", args.root] + (["--check"] if args.check else []) + (["--output", args.output] if args.output else [])
        return gen_sources.main(argv)
    main, subcommand, arg_name = _FORWARDS[args.command]
    return main([subcommand, getattr(args, arg_name)])


def main(argv: Sequence[str] | None = None) -> int:
    return route(build_parser().parse_args(argv))


if __name__ == "__main__":
    sys.exit(main())
