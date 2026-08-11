package xpath

import (
	antchfx "github.com/freemed/xpath"
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
// Returns nil if the expression cannot be compiled (including on parser panics).
func Compile(path string) (expr *Expression) {
	defer func() {
		if r := recover(); r != nil {
			expr = nil
		}
	}()
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

// CompileWithResolvers compiles an XPath expression with variable and function
// resolvers. Returns nil if compilation fails.
func CompileWithResolvers(path string, namespaces map[string]string,
	varResolver antchfx.VariableResolver, funcResolver antchfx.FunctionResolver) (expr *Expression) {
	defer func() {
		if r := recover(); r != nil {
			expr = nil
		}
	}()
	e, err := antchfx.CompileWithResolvers(path, namespaces, varResolver, funcResolver)
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
