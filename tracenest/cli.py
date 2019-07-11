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


