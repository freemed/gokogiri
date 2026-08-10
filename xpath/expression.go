package xpath

import "errors"

// Expression is a compiled XPath expression.
type Expression struct {
	xpath string
}

// Check validates an XPath expression syntax.
func Check(path string) (err error) {
	if len(path) == 0 {
		return errors.New("empty xpath expression")
	}
	// Phase 4: use antchfx/xpath for validation
	return nil
}

// Compile compiles an XPath expression string.
func Compile(path string) (expr *Expression) {
	if len(path) == 0 {
		return
	}
	// Phase 4: use antchfx/xpath for compilation
	return &Expression{xpath: path}
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
	// Phase 4
}
