# Usage & API

The public API lives at the module root (`github.com/go-ruby-set/set`). It is **Ruby-shaped but Go-idiomatic**: the method names mirror `Set`'s (`add?` → `AddQ`, `subset?` → `SubsetQ`, `^` → `XorSym`), while the surface follows Go conventions — value types, no global state, an explicit element-identity seam.

!!! success "Status: implemented"
    The library is built and importable as `github.com/go-ruby-set/set`, bound into
    `rbgo` as a native module; see [Roadmap](roadmap.md).

## Install

```sh
go get github.com/go-ruby-set/set
```

## Element identity — the Hasher seam

In MRI a `Set` keys its members by Ruby's `hash` / `eql?` protocol. Go cannot
know those semantics, so the host supplies them through a **`Hasher`** — a
function mapping a member to the comparable Go key under which two members are
considered equal:

```go
type Hasher func(elem any) any

s1 := set.New(1, 2, 3)                       // identity-keyed (comparable members)
byContent := func(e any) any { return fmt.Sprintf("%v", e) }
s2 := set.NewWith(byContent, "a", "a", "b")  // host hash/eql? — the two "a"s coincide; Size()==2
```

## Worked example

```go
a := set.New(1, 2, 3, 4)
b := set.New(3, 4, 5, 6)

a.Union(b)        // Set[1, 2, 3, 4, 5, 6]
a.Intersection(b) // Set[3, 4]
a.Difference(b)   // Set[1, 2]
a.XorSym(b)       // Set[1, 2, 5, 6]
a.SubsetQ(set.New(1, 2, 3, 4, 5)) // true

// classify: Hash{ block-value => Set }
for _, g := range set.New(1, 2, 3, 4, 5, 6).
    Classify(func(e any) any { return e.(int) % 3 }).Groups() {
    fmt.Printf("%v => %v\n", g.Value, g.Set)
}
```

## Surface

```go
func New(elems ...any) *Set               // identity-keyed (comparable members)
func NewWith(h Hasher, elems ...any) *Set // host hash/eql? semantics

// membership & mutation
func (s *Set) Add(elem any) *Set; func (s *Set) AddQ(elem any) bool
func (s *Set) Delete(elem any) *Set; func (s *Set) DeleteQ(elem any) bool
func (s *Set) Include(elem any) bool; func (s *Set) Size() int; func (s *Set) Empty() bool
func (s *Set) Clear() *Set; func (s *Set) Each(fn func(any) error) error
func (s *Set) ToSlice() []any; func (s *Set) Dup() *Set
func (s *Set) Merge(others ...*Set) *Set; func (s *Set) Subtract(other *Set) *Set

// set algebra
func (s *Set) Union(other *Set) *Set; func (s *Set) Intersection(other *Set) *Set
func (s *Set) Difference(other *Set) *Set; func (s *Set) XorSym(other *Set) *Set
func (s *Set) SubsetQ(other *Set) bool; func (s *Set) ProperSubsetQ(other *Set) bool
func (s *Set) SupersetQ(other *Set) bool; func (s *Set) ProperSupersetQ(other *Set) bool
func (s *Set) DisjointQ(other *Set) bool; func (s *Set) IntersectQ(other *Set) bool
func (s *Set) EqualQ(other *Set) bool

// enumeration & higher-order
func (s *Set) Map(fn func(any) any) []any; func (s *Set) Select(fn func(any) bool) *Set
func (s *Set) Reject(fn func(any) bool) *Set; func (s *Set) CollectBang(fn func(any) any) *Set
func (s *Set) Classify(fn func(any) any) *ClassifyResult
func (s *Set) GroupBy(fn func(any) any) *GroupByResult
func (s *Set) Divide(fn func(any) any) []*Set; func (s *Set) DivideRel(rel func(a, b any) bool) []*Set
func (s *Set) FlattenSet() *Set; func (s *Set) SortedSlice(less func(a, b any) bool) []any
func (s *Set) Inspect(stringFn func(any) string) string; func (s *Set) String() string
```

`Map` / `Collect` and `SortedSlice` return slices, matching MRI (`Set#map` and
`Set#sort` return an `Array`, not a `Set`). MRI 4.0 dropped `SortedSet` from
core, so there is no `SortedSet` type; the faithful equivalent of `Set#sort` is
`SortedSlice`.

## MRI conformance

A **differential oracle** runs set algebra, the subset/superset/disjoint
predicates, `classify`, both `divide` forms and the `Set[…]` inspect through both
the system `ruby` and this library and compares them **byte-for-byte** — gated on
`RUBY_VERSION >= "4.0"` (where `Set` is core and renders as `Set[…]`), and
skipping itself where `ruby` is absent so the cross-arch lanes still validate the
library.
