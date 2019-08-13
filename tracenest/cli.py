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
