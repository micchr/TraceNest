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
        "summary": analysis.to_dict(),
        "spans": [n.span.to_dict() for n in tree if n.span_id != tree.synthetic_root_id],
    }


def render_json(tree: SpanTree, analysis: TraceAnalysis, *, indent: int = 2) -> str:
    """Render a normalized JSON report string."""
    return json.dumps(_report_dict(tree, analysis), indent=indent, sort_keys=False)


def _fmt_ms(ms: float) -> str:
    if ms >= 1000:
        return f"{ms / 1000:.2f}s"
    return f"{ms:.1f}ms"


def _bars(tree: SpanTree, analysis: TraceAnalysis) -> str:
    """Render the nested waterfall rows as HTML."""
    total = analysis.total_ms or 1.0
    origin = min((r.span.start_ms for r in tree.roots), default=0.0)
    crit = set(analysis.critical_path)
    rows: List[str] = []

    for node in tree:
        span = node.span
        if span.span_id == tree.synthetic_root_id:
            continue
        left = (span.start_ms - origin) / total * 100.0
        width = max(0.6, span.duration_ms / total * 100.0)
        color = _KIND_COLORS.get(span.kind.value, _KIND_COLORS["unknown"])
        indent = node.depth * 16
        classes = ["row"]
        if span.span_id in crit:
            classes.append("crit")
        if span.status is SpanStatus.ERROR:
            classes.append("err")
        label = html.escape(span.name)
        meta = f"{span.kind.value} · {_fmt_ms(span.duration_ms)}"
        rows.append(
            f'<div class="{" ".join(classes)}" style="padding-left:{indent}px">'
            f'<div class="lbl" title="{label}">{label}'
            f'<span class="sub">{html.escape(meta)}</span></div>'
            f'<div class="track"><div class="bar" '
            f'style="left:{left:.3f}%;width:{width:.3f}%;background:{color}"></div></div>'
            f"</div>"
        )
    return "\n".join(rows)


def _legend() -> str:
    items = []
    for kind, color in _KIND_COLORS.items():
        if kind == "unknown":
            continue
        items.append(
            f'<span class="chip"><i style="background:{color}"></i>{kind}</span>'
        )
    return "".join(items)


def _summary_cards(analysis: TraceAnalysis) -> str:
    cards = [
        ("total", _fmt_ms(analysis.total_ms)),
        ("spans", str(analysis.span_count)),
        ("depth", str(analysis.max_depth)),
        ("critical path", _fmt_ms(analysis.critical_path_ms)),
        ("errors", str(len(analysis.error_span_ids))),
        ("retry groups", str(len(analysis.retries))),
    ]
    return "".join(
        f'<div class="card"><div class="k">{html.escape(k)}</div>'
        f'<div class="v">{html.escape(v)}</div></div>'
        for k, v in cards
    )


def _retry_rows(analysis: TraceAnalysis) -> str:
    if not analysis.retries:
        return '<tr><td colspan="4" class="muted">no retry clusters detected</td></tr>'
    out = []
    for r in analysis.retries:
        state = "recovered" if r.recovered else ("failing" if r.errors else "repeated")
        out.append(
            f"<tr><td>{html.escape(r.name)}</td><td>{r.attempts}</td>"
            f"<td>{r.errors}</td><td>{state}</td></tr>"
        )
    return "".join(out)


def render_html(tree: SpanTree, analysis: TraceAnalysis, *, title: str = "tracenest report") -> str:
    """Render a single self-contained HTML report."""
    payload = json.dumps(_report_dict(tree, analysis))
