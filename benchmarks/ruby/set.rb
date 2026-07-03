# frozen_string_literal: true
# SPDX-License-Identifier: BSD-3-Clause
require "set"
require_relative "_harness"
n = 1000
as = (0...n).to_a
bs = (n/2...(n + n/2)).to_a          # 50% overlap
a = Set.new(as)
b = Set.new(bs)
bench("build-1000",        500) { Set.new(as) }
bench("union-1000",        500) { a | b }
bench("intersection-1000", 500) { a & b }
bench("membership-1000",   500) { as.each { |x| a.include?(x) } }
