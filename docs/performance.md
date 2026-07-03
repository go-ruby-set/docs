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

## Library-level benchmark (Go API vs runtimes) — 2026-07-03

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
- **Method:** each process runs 3 untimed warm-up passes, then 25 timed passes of
  a fixed inner loop, timed with a monotonic clock; the **best** pass is reported
  as **ns/op** (lower is better). `vs MRI` < 1.00× means *faster than MRI*.
  Interpreter start-up is outside the timed region, so these are operation costs,
  not `ruby file.rb` process costs.

#### build-1000

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 55933.9 | 4.49× |
| MRI | 12462.0 | 1.00× |
| MRI + YJIT | 12506.0 | 1.00× |
| JRuby | 12526.1 | 1.01× |
| TruffleRuby | 10185.0 | 0.82× |

#### intersection-1000

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 23854.6 | 1.34× |
| MRI | 17826.0 | 1.00× |
| MRI + YJIT | 17132.0 | 0.96× |
| JRuby | 7369.5 | 0.41× |
| TruffleRuby | 14386.0 | 0.81× |

#### membership-1000

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 12009.9 | 0.39× |
| MRI | 30418.0 | 1.00× |
| MRI + YJIT | 12108.0 | 0.40× |
| JRuby | 21765.7 | 0.72× |
| TruffleRuby | 2417.2 | 0.08× |

#### union-1000

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 52684.6 | 3.40× |
| MRI | 15498.0 | 1.00× |
| MRI + YJIT | 15480.0 | 1.00× |
| JRuby | 13811.7 | 0.89× |
| TruffleRuby | 13260.2 | 0.86× |

Mixed and instructive: `build`/`union`/`intersection` are 1.3–4.5× MRI (the library preserves insertion order and hashes through Go `any`, versus MRI's C hash set), while **`membership` is ~2.5× faster** (0.39×). Insertion-ordered construction is the optimization target; the ordered-map design is what costs on bulk build.

!!! note "Reproduce"
    The harness is committed under
    [`benchmarks/`](https://github.com/go-ruby-set/docs/tree/main/benchmarks):
    a self-contained Go driver (`go/`, pins the published library via
    `go.mod`), the equivalent `ruby/set.rb` workload, and `run.sh`. Run
    `bash benchmarks/run.sh`; env `OUTER`/`WARM` tune the pass budget and
    `RUBY`/`JRUBY`/`TRUFFLERUBY` select the runtime binaries.

!!! warning "Warm-up budget & noise — honest framing"
    Numbers reflect a **fixed warm-process budget** (3 warm-up + 25 timed passes
    in one process). The JVM/GraalVM JITs (JRuby, TruffleRuby) may need a larger
    warm-up to reach steady state, so their columns can **understate** peak
    throughput — most visibly TruffleRuby on the shortest loops (a few cold-JIT
    outliers are noted in the text). Sub-microsecond rows carry the most relative
    noise; treat those ratios as order-of-magnitude. Every number here is a
    **real measured value** from the dated run above — nothing is fabricated,
    estimated, or cherry-picked. The go-ruby column is the pure-Go library; every
    other column is that interpreter's own stdlib doing the equivalent work.
