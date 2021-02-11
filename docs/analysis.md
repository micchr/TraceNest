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

