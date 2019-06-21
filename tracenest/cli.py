"""Command-line interface for tracenest.

Usage
-----
    python -m tracenest analyze  TRACE [--json]
    python -m tracenest tree     TRACE
    python -m tracenest report   TRACE [--out FILE] [--format html|json]

``TRACE`` may be a path or ``-`` for standard input.
"""

from __future__ import annotations

