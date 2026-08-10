package xpath

import (
	"errors"
	"strconv"

	antchfx "github.com/antchfx/xpath"
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

// AttrNode is a lightweight attribute node returned by XPath attribute queries.
// It implements NodeAdapter so it can be collected as a nodeset result.
type AttrNode struct {
	Name_         string
	Value_        string
	Prefix_       string
	NamespaceURI_ string
}

func (a *AttrNode) XPathType() int              { return 2 } // AttributeNode
func (a *AttrNode) XPathName() string            { return a.Name_ }
func (a *AttrNode) XPathValue() string           { return a.Value_ }
func (a *AttrNode) XPathPrefix() string          { return a.Prefix_ }
func (a *AttrNode) XPathNamespaceURI() string    { return a.NamespaceURI_ }
func (a *AttrNode) XPathParent() interface{}     { return nil }
func (a *AttrNode) XPathFirstChild() interface{}  { return nil }
func (a *AttrNode) XPathNextSibling() interface{} { return nil }
func (a *AttrNode) XPathPrevSibling() interface{} { return nil }
func (a *AttrNode) XPathAttributes() interface{}  { return nil }
func (a *AttrNode) XPathCopy() interface{}        { return a }

// XPath is the XPath evaluation context.
type XPath struct {
	resultType   XPathObjectType
	resultNodes  []interface{}
	resultString string
	resultNumber float64
	resultBool   bool
	namespaces   map[string]string
	resolver     VariableScope
	position     int
	size         int
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
	return &XPath{
		namespaces: make(map[string]string),
	}
}

// RegisterNamespace registers a namespace prefix/URI pair.
func (xpath *XPath) RegisterNamespace(prefix, href string) bool {
	xpath.namespaces[prefix] = href
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
	if nodePtr == nil || xpathExpr == nil || xpathExpr.expr == nil {
		return errors.New("nil node or expression in xpath evaluate")
	}

	// Get a navigator from the node
	adapter, ok := nodePtr.(NodeAdapter)
	if !ok {
		return errors.New("node does not implement NodeAdapter")
	}

	nav := NewNavigator(adapter)

	// Evaluate with antchfx/xpath
	result := xpathExpr.expr.Evaluate(nav)

	// Store result based on type
	switch v := result.(type) {
	case *antchfx.NodeIterator:
		xpath.resultType = XPATH_NODESET
		xpath.resultNodes = nil
		// Collect all nodes
		for v.MoveNext() {
			current := v.Current()
			if nn, ok := current.(*nodeNavigator); ok {
				xpath.resultNodes = append(xpath.resultNodes, nn.resultNode())
			}
		}
	case bool:
		xpath.resultType = XPATH_BOOLEAN
		xpath.resultBool = v
	case float64:
		xpath.resultType = XPATH_NUMBER
		xpath.resultNumber = v
	case string:
		xpath.resultType = XPATH_STRING
		xpath.resultString = v
	default:
		// Fallback: try NodeIterator via Select
		iter := xpathExpr.expr.Select(nav)
		xpath.resultType = XPATH_NODESET
		xpath.resultNodes = nil
		for iter.MoveNext() {
			current := iter.Current()
			if nn, ok := current.(*nodeNavigator); ok {
				xpath.resultNodes = append(xpath.resultNodes, nn.resultNode())
			}
		}
	}

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
	switch xpath.resultType {
	case XPATH_STRING:
		return xpath.resultString, nil
	case XPATH_NUMBER:
		return strconv.FormatFloat(xpath.resultNumber, 'f', -1, 64), nil
	case XPATH_BOOLEAN:
		if xpath.resultBool {
			return "true", nil
		}
		return "false", nil
	case XPATH_NODESET:
		if len(xpath.resultNodes) > 0 {
			if adapter, ok := xpath.resultNodes[0].(NodeAdapter); ok {
				return adapter.XPathValue(), nil
			}
		}
		return "", nil
	}
	return "", errors.New("not a string result")
}

// ResultAsNumber coerces the result to a number.
func (xpath *XPath) ResultAsNumber() (val float64, err error) {
	return xpath.resultNumber, nil
}

// ResultAsBoolean coerces the result to a boolean.
// In XPath 1.0:
//   - A nodeset is true if it is non-empty.
//   - A string is true if it is non-empty.
//   - A number is true if it is non-zero and not NaN.
func (xpath *XPath) ResultAsBoolean() (val bool, err error) {
	switch xpath.resultType {
	case XPATH_BOOLEAN:
		return xpath.resultBool, nil
	case XPATH_NODESET:
		return len(xpath.resultNodes) > 0, nil
	case XPATH_STRING:
		return xpath.resultString != "", nil
	case XPATH_NUMBER:
		return xpath.resultNumber != 0, nil
	}
	return false, nil
}

// SetResolver attaches a variable/function resolver.
func (xpath *XPath) SetResolver(v VariableScope) {
	xpath.resolver = v
}

// SetContextPosition sets the position/size for position()/last().
func (xpath *XPath) SetContextPosition(position, size int) {
	xpath.position = position
	xpath.size = size
}

// GetContextPosition returns the current position/size.
func (xpath *XPath) GetContextPosition() (position, size int) {
	return xpath.position, xpath.size
}

// Free releases resources.
func (xpath *XPath) Free() {
	xpath.resultNodes = nil
}

// XPathObjectToValue converts a raw XPath result to a Go value.
func XPathObjectToValue(obj interface{}) (result interface{}) {
	return obj
}
