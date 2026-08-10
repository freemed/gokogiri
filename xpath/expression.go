package xpath

import (
	antchfx "github.com/antchfx/xpath"
)

// Expression is a compiled XPath expression.
type Expression struct {
	expr   *antchfx.Expr
	xpath  string
}

// Check validates an XPath expression syntax.
func Check(path string) (err error) {
	_, err = antchfx.Compile(path)
	return
}

// Compile compiles an XPath expression string.
func Compile(path string) (expr *Expression) {
	e, err := antchfx.Compile(path)
	if err != nil {
		return nil
	}
	return &Expression{expr: e, xpath: path}
}

// CompileWithNS compiles with namespace bindings.
func CompileWithNS(path string, namespaces map[string]string) (expr *Expression) {
	e, err := antchfx.CompileWithNS(path, namespaces)
	if err != nil {
		return nil
	}
	return &Expression{expr: e, xpath: path}
}

// String returns the original XPath string.
func (exp *Expression) String() string {
	if exp == nil {
		return ""
	}
	return exp.xpath
}

// Free releases compilation resources.
func (exp *Expression) Free() {
	exp.expr = nil
}
