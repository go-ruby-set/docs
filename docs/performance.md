# Performance

`go-ruby-set/set` is the pure-Go library that
[`rbgo`](https://github.com/go-embedded-ruby/ruby) binds for Ruby's `set`. This
page records a **comparative benchmark** of that module against the reference
Ruby runtimes, part of the ecosystem-wide per-module parity suite.

## What is measured

The **same** Ruby script — the same set-algebra workload — building two sets and computing their `|` `&` `-` `^` plus subset/classify — is run under every runtime. `rbgo`'s
number reflects **this pure-Go library doing the work**; every other column is
that interpreter's own `set` implementation. So the comparison is the
**Ruby-visible operation**, apples-to-apples across interpreters. The script
prints a deterministic checksum and its output is checked **byte-identical to
MRI** before timing.

- **Method:** best-of-N wall time (best, not mean, to suppress scheduler noise);
  single-shot processes, no warm-up beyond the script's own loop.
- **Runtimes:** `ruby` (MRI, the oracle) and `ruby --yjit`; `jruby` (on the JVM);
  `truffleruby` (GraalVM). JRuby and TruffleRuby are timed **cold, single-shot**,
  so they carry JVM / Graal startup on every run.
- The benchmark script and harness live in rbgo's repo under
  [`bench/modules/`](https://github.com/go-embedded-ruby/ruby/tree/main/bench/modules)
  (`set.rb` + `run.sh`). Reproduce:
  `RBGO=./rbgo TRUFFLE=truffleruby bash bench/modules/run.sh 5`.

## Result (best of 5, ms)

| Runtime | time | vs MRI |
| --- | ---: | ---: |
| **rbgo** (go-ruby-set) | 2220 | 10.09× |
| MRI (ruby 4.0.5) | 220 | 1.00× |
| MRI + YJIT | 220 | 1.00× |
| JRuby 10.1.0.0 | 1240 | 5.64× |
| TruffleRuby 34.0.1 | 420 | 1.91× |

rbgo runs on **go-ruby-set** and is **~10x slower than MRI** here (10.09x): the set-algebra loop drives a very high rate of per-element interpreter dispatch through Set's Ruby methods, which is rbgo's most expensive primitive (frame setup + interface dispatch per send). This is the top per-module optimization target for go-ruby-set; output stays byte-identical to MRI.

!!! note "Honest framing"
    JRuby and TruffleRuby are timed **cold, single-shot**, so they carry JVM /
    Graal startup on every run — read them as one-shot `ruby file.rb` costs, the
    same way `rbgo` and MRI are measured, not as steady-state JIT numbers. Rows
    that complete in well under ~200 ms carry the most relative noise; treat
    their ratios as order-of-magnitude. These are **real measured numbers** from
    the 2026-06-30 run (Apple M-series; `ruby 4.0.5 +PRISM`, `jruby 10.1.0.0`,
    `truffleruby 34.0.1`) — nothing is fabricated or cherry-picked.

## Library-level benchmark (Go API vs runtimes) — 2026-07-03 (specialized hash)

This section measures the **pure-Go library directly, through its Go API** — not
the `rbgo` interpreter path recorded above. It isolates the library primitive
from Ruby-interpreter dispatch, answering the parity question head-on: *is the
pure-Go implementation as fast as the reference runtime's own `set`?* The
**same workload, same inputs, same iteration counts** run through the Go library
and through each reference runtime's stdlib; outputs were checked identical to
MRI before any timing.

- **Host:** Apple M4 Max (`Mac16,5`, arm64), macOS 26.5.1 — **date 2026-07-03**.
- **Runtimes:** Go 1.26.4 · MRI `ruby 4.0.5 +PRISM` · MRI + YJIT · JRuby 10.1.0.0
  (OpenJDK 25) · TruffleRuby 34.0.1 (GraalVM CE Native).
- **Method:** each process runs 5 untimed warm-up passes, then 60 timed passes of
  a fixed inner loop, timed with a monotonic clock; the **best** pass is reported
  as **ns/op** (lower is better). `vs MRI` < 1.00× means *faster than MRI*.
  Interpreter start-up is outside the timed region, so these are operation costs,
  not `ruby file.rb` process costs.

### What changed (specialized-hash membership index)

The previous rework moved the members into two parallel insertion-ordered slices
(`keys`, `members`) with a membership-only index — but that index was a single
`map[any]struct{}`. A generic-`any` map hashes **every** key through the
interface's type descriptor, roughly **twice** the cost of a map keyed by a
concrete comparable type, and that interface hashing was the dominant residual
cost on `build` and `union` — the two ops still above MRI/YJIT.

Most real Sets are homogeneous (all Integer, or all String / Symbol), so the
membership index now keeps a **concrete-typed map** for the common key types —
all Go `int`, all `int64`, all `string` — and falls back to the generic
`map[any]` only for mixed or other-typed sets. A small `membership` type carries a
`kind` discriminator plus the active map; the hot paths (`has` / `add` / `putNew`
/ `del`) are one predictable branch on `kind` followed by a concrete-typed map
op, and `clone()` uses `maps.Clone` on the active typed map so `Union` keeps its
fast bulk copy. A micro-benchmark confirms the premise directly: building a
1000-key `map[int]` is ~1.8× faster and probing it ~1.9× faster than the
`map[any]` equivalent.

**Exact Ruby semantics are preserved.** Each fast path is gated on **one**
concrete type, so `int(1)`, `int64(1)`, `1.0` and `"1"` stay four distinct
members: the first key of a foreign concrete type promotes the whole index to the
generic map, re-boxing every existing key under its **original** type so nothing
is lost or coalesced. Insertion-order iteration, dedup, delete, the full
predicate/algebra surface, and the **MRI oracle** are unchanged — including a test
for a set that starts homogeneous and then gains a mixed-type member.

Before → after `vs MRI` (before = the prior bulk-path result; same host, harness):

| Op | before | after (vs MRI) | after (vs YJIT) | verdict |
| --- | ---: | ---: | ---: | --- |
| build-1000 | 1.87× | **1.35×** | 1.33× | honest floor (order slices) |
| union-1000 | 1.57× | **1.07×** | 1.06× | parity |
| intersection-1000 | 1.11× | **0.78×** | 0.82× | **beats MRI + YJIT** |
| membership-1000 | 0.41× | **0.28×** | 0.74× | **beats MRI + YJIT** |

#### build-1000

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 16312.8 | 1.35× |
| MRI | 12054.0 | 1.00× |
| MRI + YJIT | 12300.0 | 1.02× |
| JRuby | 11710.2 | 0.97× |
| TruffleRuby | 9649.0 | 0.80× |

#### intersection-1000

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 13500.2 | 0.78× |
| MRI | 17228.0 | 1.00× |
| MRI + YJIT | 16454.0 | 0.96× |
| JRuby | 7125.0 | 0.41× |
| TruffleRuby | 13555.2 | 0.79× |

#### membership-1000

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 8511.9 | 0.28× |
| MRI | 29926.0 | 1.00× |
| MRI + YJIT | 11504.0 | 0.38× |
| JRuby | 18216.5 | 0.61× |
| TruffleRuby | 2517.8 | 0.08× |

#### union-1000

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 16546.4 | 1.07× |
| MRI | 15444.0 | 1.00× |
| MRI + YJIT | 15560.0 | 1.01× |
| JRuby | 13324.7 | 0.86× |
| TruffleRuby | 13053.3 | 0.85× |

After specializing the index, **`intersection` (0.78×) and `membership` (0.28×)
now beat both MRI and YJIT**, `union` is at **parity** (1.07× / 1.06× vs YJIT —
within run-to-run noise), and `build` closes to **1.35×** (was 1.87×).

!!! note "The honest floor on build / union"
    `build` and `union` stay just above 1× for a **structural** reason, not
    redundant work — and it is well profiled. The typed-map dedup **on its own**
    is already ~MRI parity (a pre-sized `map[int64]` read+insert of 1000 keys
    measures ~12.8 µs, next to MRI's ~12.1 µs `build`). The residual ~0.3× is the
    **two parallel insertion-order slices** Go must maintain because Go's `map`
    is *not* insertion-ordered — bookkeeping MRI gets for free from its natively
    insertion-ordered `Hash`. Closing it further would mean building a custom
    insertion-ordered hash table (re-implementing what the slices + map already
    are) for a sub-microsecond-per-op gain; the generic-`any` hashing that used to
    dominate is now gone. This is the honest wall for these two ops.

!!! note "Reproduce"
    The harness is committed under
    [`benchmarks/`](https://github.com/go-ruby-set/docs/tree/main/benchmarks):
    a self-contained Go driver (`go/`, which pins this library's feature commit
    via `go.mod`), the equivalent `ruby/set.rb` workload, and `run.sh`. Run
    `OUTER=60 WARM=5 bash benchmarks/run.sh`; env `OUTER`/`WARM` tune the pass
    budget and `RUBY`/`JRUBY`/`TRUFFLERUBY` select the runtime binaries.

!!! warning "Warm-up budget & noise — honest framing"
    Numbers reflect a **fixed warm-process budget** (5 warm-up + 60 timed passes
    in one process, best pass reported). The JVM/GraalVM JITs (JRuby, TruffleRuby)
    may need a larger warm-up to reach steady state, so their columns can
    **understate** peak throughput — most visibly TruffleRuby on the shortest
    loops. Sub-microsecond rows carry the most relative noise; treat those ratios
    as order-of-magnitude, and read `union` vs YJIT (1.06×) as parity. Every
    number here is a **real measured value** from the dated run above — nothing is
    fabricated, estimated, or cherry-picked. The go-ruby column is the pure-Go
    library; every other column is that interpreter's own stdlib doing the
    equivalent work.
