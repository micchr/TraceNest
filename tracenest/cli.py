"""Command-line interface for tracenest.

Usage
-----
    python -m tracenest analyze  TRACE [--json]
    python -m tracenest tree     TRACE
    python -m tracenest report   TRACE [--out FILE] [--format html|json]

``TRACE`` may be a path or ``-`` for standard input.
"""

from __future__ import annotations

import argparse
import sys
from pathlib import Path
from typing import List, Optional

from .analyze import analyze
from .parser import ParseError, load_spans, parse_spans
from .report import render_html, render_json
from .tree import SpanTree, build_tree


def _read(source: str) -> SpanTree:
    if source == "-":
        spans = parse_spans(sys.stdin.read())
    else:
        spans = load_spans(source)
    return build_tree(spans)


def _cmd_analyze(args: argparse.Namespace) -> int:
    tree = _read(args.trace)
    result = analyze(tree)
    if args.json:
        print(render_json(tree, result))
        return 0
    print(_text_summary(result))
    return 0


def _text_summary(result) -> str:
    lines: List[str] = []
    lines.append(f"spans        : {result.span_count}")
    lines.append(f"total        : {result.total_ms:.1f} ms")
    lines.append(f"max depth    : {result.max_depth}")
    lines.append(f"critical path: {result.critical_path_ms:.1f} ms "
                 f"({len(result.critical_path)} spans)")
    lines.append("")
    lines.append("by kind:")
    for kind, ms in sorted(result.kind_ms.items(), key=lambda kv: -kv[1]):
        count = result.kind_counts.get(kind, 0)
        lines.append(f"  {kind:<10} {ms:>10.1f} ms  ({count})")
    lines.append("")
    lines.append("status:")
    for status, count in sorted(result.status_counts.items()):
        lines.append(f"  {status:<10} {count}")
    if result.retries:
        lines.append("")
        lines.append("retries:")
        for r in result.retries:
            tag = "recovered" if r.recovered else ("failing" if r.errors else "repeated")
            lines.append(f"  {r.name:<20} attempts={r.attempts} errors={r.errors} [{tag}]")
    return "\n".join(lines)


def _cmd_tree(args: argparse.Namespace) -> int:
    tree = _read(args.trace)
    for node in tree:
        if node.span_id == tree.synthetic_root_id:
            marker = "*"
        else:
            marker = {"error": "x", "ok": "+", "cancelled": "~"}.get(
                node.span.status.value, " "
            )
        indent = "  " * node.depth
        dur = node.span.duration_ms
        print(f"{marker} {indent}{node.span.name} "
              f"[{node.span.kind.value}] {dur:.1f}ms")
    return 0


def _cmd_report(args: argparse.Namespace) -> int:
    tree = _read(args.trace)
    result = analyze(tree)
    if args.format == "json":
        text = render_json(tree, result)
    else:
        text = render_html(tree, result)
    if args.out:
        Path(args.out).write_text(text, encoding="utf-8")
        print(f"wrote {args.out} ({len(text)} bytes)")
    else:
        sys.stdout.write(text)
    return 0


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="tracenest",
        description="agent trace explorer: nest spans, find the critical path, render reports",
    )
    sub = parser.add_subparsers(dest="command", required=True)

    p_an = sub.add_parser("analyze", help="print timing and retry summaries")
    p_an.add_argument("trace", help="trace file path or - for stdin")
    p_an.add_argument("--json", action="store_true", help="emit JSON instead of text")
    p_an.set_defaults(func=_cmd_analyze)

    p_tr = sub.add_parser("tree", help="print the nested span tree")
    p_tr.add_argument("trace", help="trace file path or - for stdin")
    p_tr.set_defaults(func=_cmd_tree)

    p_re = sub.add_parser("report", help="render a self-contained HTML or JSON report")
    p_re.add_argument("trace", help="trace file path or - for stdin")
    p_re.add_argument("--out", help="output file (default: stdout)")
    p_re.add_argument("--format", choices=("html", "json"), default="html")
    p_re.set_defaults(func=_cmd_report)

    return parser


