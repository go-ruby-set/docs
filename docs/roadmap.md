# Roadmap

`go-ruby-set/set` is grown **test-first**, each capability differential-tested against
MRI rather than built in isolation. The deterministic, interpreter-independent
slice of Ruby's `Set` extracted from rbgo's internals is **complete**.

| Stage | What | Status |
| --- | --- | --- |
| Core & membership | Insertion-ordered storage with `add` / `add?`, `delete` / `delete?`, `include?`, `size`, `empty?`, `clear`, `each`, `to_a`, `dup`, `merge`, `subtract`. | **Done** |
| Set algebra | `Union` (`|`), `Intersection` (`&`), `Difference` (`-`) and `XorSym` (`^`), each producing a fresh, insertion-ordered set. | **Done** |
| Predicates | `subset?`, `proper_subset?`, `superset?`, `proper_superset?`, `disjoint?`, `intersect?`, `==`. | **Done** |
| Enumeration & higher-order | `map` / `select` / `reject` / `collect!`, `classify`, `group_by`, both `divide` forms, recursive `flatten`, `sort` via `SortedSlice`. | **Done** |
| Element identity (Hasher) | Pluggable `Hasher` for the host's `hash` / `eql?`; identity keying for plain Go data via `New`. | **Done** |
| Differential oracle & coverage | Algebra, predicates, `classify`, `divide`, `Set[…]` inspect checked byte-for-byte against `ruby` (≥ 4.0); 100% coverage, green across 6 arches and 3 OSes. | **Done** |

## Documented out-of-scope boundaries

These are **deliberate**, recorded so the module's surface is unambiguous:

- **No interpreter.** The library implements the deterministic algorithm; it
  never runs arbitrary Ruby. Anything that needs a live binding or evaluation is
  the consumer's job — that is why `rbgo` binds this module rather than the
  reverse.
- **Reference is reference Ruby (MRI).** Byte-for-byte conformance targets MRI's
  behaviour; differences across MRI releases are matched to the reference used by
  the differential oracle.
- **Standalone & reusable.** The module has no dependency on the Ruby runtime;
  the dependency runs the other way.

See [Usage & API](api.md) for the surface and [Why pure Go](why.md) for the
deterministic/interpreter split.
