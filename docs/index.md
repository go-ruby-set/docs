# go-ruby-set documentation

**Ruby's Set — the unordered, unique collection with full set algebra — in pure Go, no cgo.**

`go-ruby-set/set` is a faithful, pure-Go (zero cgo) reimplementation of Ruby's `Set`,
matching reference Ruby (MRI) byte-for-byte. The module path is
`github.com/go-ruby-set/set`.

It is a **standalone, reusable** library importable by any Go program, and the
backend bound into [go-embedded-ruby](https://github.com/go-embedded-ruby/ruby)
by `rbgo` as a native module — the same pattern as
[go-ruby-yaml](https://github.com/go-ruby-yaml/yaml). The dependency runs the
other way: this library has **no dependency on the Ruby runtime**.

!!! success "Status: complete — MRI byte-exact"
    A faithful pure-Go port of Ruby's `Set`, validated by a **differential oracle**
    against the system `ruby` — results compared byte-for-byte — at 100%
    coverage, `gofmt` + `go vet` clean, CI green across the six 64-bit Go targets
    and three OSes.

## Quick taste

```go
a := set.New(1, 2, 3, 4)
b := set.New(3, 4, 5, 6)

a.Union(b)        // Set[1, 2, 3, 4, 5, 6]
a.Intersection(b) // Set[3, 4]
a.Difference(b)   // Set[1, 2]
a.XorSym(b)       // Set[1, 2, 5, 6]
a.SubsetQ(set.New(1, 2, 3, 4, 5)) // true
```

## Repositories

| Repo | What it is |
| --- | --- |
| [`set`](https://github.com/go-ruby-set/set) | the library — Set in pure Go |
| [`docs`](https://github.com/go-ruby-set/docs) | this documentation site (MkDocs Material, versioned with mike) |
| [`go-ruby-set.github.io`](https://github.com/go-ruby-set/go-ruby-set.github.io) | the organization landing page (Hugo) |
| [`brand`](https://github.com/go-ruby-set/brand) | logo and brand assets |

## Principles

- **Pure Go, `CGO_ENABLED=0`** — trivial cross-compilation, a single static
  binary, no C toolchain.
- **MRI byte-exact.** Output matches reference Ruby exactly, not approximately,
  validated by a differential oracle against the `ruby` binary.
- **Standalone & reusable.** Extracted from rbgo's internals; no dependency on
  the Ruby runtime — the dependency runs the other way.
- **100% test coverage** is the target, enforced as a CI gate, across 6 arches
  and 3 OSes.

## Where to go next

- [Why pure Go](why.md) — why this slice of Ruby is deterministic enough to live
  as a standalone, interpreter-independent Go library.
- [Usage & API](api.md) — the public surface and worked examples.
- [Roadmap](roadmap.md) — what is done and what is downstream by design.

Source lives at [github.com/go-ruby-set/set](https://github.com/go-ruby-set/set).
