// Original work Copyright (c) 2016 Jonas Obrist (https://github.com/ojii/gettext.go)
// Modified work Copyright (c) 2018 DeineAgentur UG https://www.deineagentur.com
// Modified work Copyright (c) 2018-present gotext maintainers (https://github.com/ibitcat/gotext)
//
// Licensed under the 3-Clause BSD License. See LICENSE in the project root for license information.

package plurals

// Expression is a plurals expression. Eval evaluates the expression for
// a given n value. Use plurals.Compile to generate Expression instances.
type Expression interface {
	Eval(n uint32) int
}

type constValue struct {
	value int
}

func (c constValue) Eval(n uint32) int {
	return c.value
}

type test interface {
	test(n uint32) bool
}

type ternary struct {
	test      test
	trueExpr  Expression
	falseExpr Expression
}

func (t ternary) Eval(n uint32) int {
	if t.test.test(n) {
		if t.trueExpr == nil {
			return -1
		}
		return t.trueExpr.Eval(n)
	}
	if t.falseExpr == nil {
		return -1
	}
	return t.falseExpr.Eval(n)
}
