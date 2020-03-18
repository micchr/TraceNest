"""Report rendering.

Emits either a normalized JSON document or a single self-contained HTML
page (no external assets, no network) styled in a graphite/neon-green
theme. The HTML embeds an interactive nested waterfall built from inline
CSS and a small amount of vanilla JavaScript.
"""

from __future__ import annotations

import html
import json
from typing import Dict, List

from .analyze import TraceAnalysis, analyze
from .model import SpanKind, SpanStatus
from .tree import SpanTree, TreeNode

# Graphite / neon-green palette shared with the SVG assets.
_KIND_COLORS = {
    "root": "#8a94a6",
    "llm": "#39ff14",
    "tool": "#00e0a4",
    "retrieval": "#5bd6ff",
    "chain": "#c8ff5b",
    "event": "#a0f0c0",
    "unknown": "#6b7280",
}


def _report_dict(tree: SpanTree, analysis: TraceAnalysis) -> Dict[str, object]:
    return {
        "version": 1,
