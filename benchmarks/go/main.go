// SPDX-License-Identifier: BSD-3-Clause
package main

import "github.com/go-ruby-set/set"

func mk(vals []int) *set.Set {
	el := make([]any, len(vals))
	for i, v := range vals {
		el[i] = v
	}
	return set.New(el...)
}

func main() {
	n := 1000
	as := make([]int, n)
	for i := range as {
		as[i] = i
	}
	bs := make([]int, n)
	for i := range bs {
		bs[i] = n/2 + i
	}
	asAny := make([]any, n)
	for i, v := range as {
		asAny[i] = v
	}
	a := mk(as)
	b := mk(bs)
	bench("build-1000", 500, func() { sink = set.New(asAny...) })
	bench("union-1000", 500, func() { sink = a.Union(b) })
	bench("intersection-1000", 500, func() { sink = a.Intersection(b) })
	bench("membership-1000", 500, func() {
		for _, x := range as {
			sink = a.Include(x)
		}
	})
}
