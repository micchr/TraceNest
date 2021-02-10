# Analysis Reference

This note explains exactly what the numbers in a tracenest report mean.

## Self time

A span's **self time** is its own duration minus the union of its
children's intervals:

```
self(s) = duration(s) - measure( union of [child.start, child.end] )
```

Because the child intervals are merged before subtracting, overlapping
