# Analysis Reference

This note explains exactly what the numbers in a tracenest report mean.

## Self time

A span's **self time** is its own duration minus the union of its
children's intervals:

```
self(s) = duration(s) - measure( union of [child.start, child.end] )
```

Because the child intervals are merged before subtracting, overlapping
children (concurrent tool calls) are never double-counted, and self time
is clamped at zero. Self time answers "how much wall-clock time is spent
in *this* span rather than in the work it delegates?"

## Critical path

The **critical path** is the root-to-leaf chain that maximizes the sum of
self times along it. It is computed with a single post-order pass:

```
best(leaf)   = self(leaf)
best(parent) = self(parent) + max(best(child) for child in children)
```

