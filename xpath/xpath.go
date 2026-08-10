package xpath

import (
	"errors"
)

// XPathObjectType mirrors the libxml2 XPath result types.
type XPathObjectType int

const (
	XPATH_UNDEFINED   XPathObjectType = 0
	XPATH_NODESET                     = 1
	XPATH_BOOLEAN                     = 2
	XPATH_NUMBER                      = 3
	XPATH_STRING                      = 4
	XPATH_POINT                       = 5
	XPATH_RANGE                       = 6
	XPATH_LOCATIONSET                 = 7
	XPATH_USERS                       = 8
	XPATH_XSLT_TREE                   = 9
)

// XPath is the XPath evaluation context.
type XPath struct {
	resultType XPathObjectType
	resultNodes []interface{}
	resultString string
	resultNumber float64
	resultBool   bool
}

// XPathFunction is called for registered XPath extension functions.
type XPathFunction func(context VariableScope, args []interface{}) interface{}

// VariableScope knows how to resolve XPath variables and functions.
type VariableScope interface {
	ResolveVariable(string, string) interface{}
	IsFunctionRegistered(string, string) bool
	ResolveFunction(string, string) XPathFunction
}

// NewXPath creates a new XPath context from a document.
func NewXPath(docPtr interface{}) (xpath *XPath) {
	if docPtr == nil {
		return nil
	}
	return &XPath{}
}

// RegisterNamespace registers a namespace prefix/URI pair.
func (xpath *XPath) RegisterNamespace(prefix, href string) bool {
	// Phase 4 will implement with antchfx/xpath
	return true
}

// EvaluateAsNodeset evaluates an XPath expression and returns matching nodes.
func (xpath *XPath) EvaluateAsNodeset(nodePtr interface{}, xpathExpr *Expression) (nodes []interface{}, err error) {
	if nodePtr == nil {
		return
	}
	err = xpath.Evaluate(nodePtr, xpathExpr)
	if err != nil {
		return
	}
	nodes, err = xpath.ResultAsNodeset()
	return
}

// Evaluate runs an XPath expression against a context node.
func (xpath *XPath) Evaluate(nodePtr interface{}, xpathExpr *Expression) (err error) {
	if nodePtr == nil || xpathExpr == nil {
		return errors.New("nil node or expression in xpath evaluate")
	}
	// Phase 4: full antchfx/xpath evaluation
	// For now return empty
	xpath.resultType = XPATH_NODESET
	xpath.resultNodes = nil
	return nil
}

// ReturnType returns the type of the last evaluation result.
func (xpath *XPath) ReturnType() XPathObjectType {
	return xpath.resultType
}

// ResultAsNodeset returns the result as a slice of node pointers.
func (xpath *XPath) ResultAsNodeset() (nodes []interface{}, err error) {
	if xpath.resultType != XPATH_NODESET && xpath.resultType != XPATH_XSLT_TREE {
		err = errors.New("Cannot convert XPath result to nodeset")
		return
	}
	return xpath.resultNodes, nil
}

// ResultAsString coerces the result to a string.
func (xpath *XPath) ResultAsString() (val string, err error) {
	if xpath.resultType == XPATH_STRING {
		return xpath.resultString, nil
	}
	return "", errors.New("not a string result")
}

// ResultAsNumber coerces the result to a number.
func (xpath *XPath) ResultAsNumber() (val float64, err error) {
	return xpath.resultNumber, nil
}

// ResultAsBoolean coerces the result to a boolean.
func (xpath *XPath) ResultAsBoolean() (val bool, err error) {
	return xpath.resultBool, nil
}

// SetResolver attaches a variable/function resolver.
func (xpath *XPath) SetResolver(v VariableScope) {
	// Phase 4 implementation
}

// SetContextPosition sets the position/size for position()/last().
func (xpath *XPath) SetContextPosition(position, size int) {
	// Phase 4 implementation
}

// GetContextPosition returns the current position/size.
func (xpath *XPath) GetContextPosition() (position, size int) {
	return 0, 0
}

// Free releases resources.
func (xpath *XPath) Free() {
	xpath.resultNodes = nil
}

// XPathObjectToValue converts a raw XPath result to a Go value.
func XPathObjectToValue(obj interface{}) (result interface{}) {
	// Phase 4
	return nil
}
