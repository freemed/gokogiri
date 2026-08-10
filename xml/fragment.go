package xml

import (
	"bytes"
	"errors"
	"strings"
)

type DocumentFragment struct {
	Node        Node
	InEncoding  []byte
	OutEncoding []byte
}

var (
	fragmentWrapperStart = []byte("<root>")
	fragmentWrapperEnd   = []byte("</root>")
)

var ErrFailParseFragment = errors.New("failed to parse xml fragment")
var ErrEmptyFragment = errors.New("empty xml fragment")

const initChildrenNumber = 4

func parsefragment(document Document, node *XmlNode, content, url []byte, options ParseOption) (fragment *DocumentFragment, err error) {
	// Wrap content before parsing
	wrapped := append(fragmentWrapperStart, content...)
	wrapped = append(wrapped, fragmentWrapperEnd...)

	// Parse as document
	doc, parseErr := Parse(wrapped, document.InputEncoding(), url, options, document.OutputEncoding())
	if parseErr != nil {
		return nil, ErrFailParseFragment
	}
	defer doc.Free()

	// The root element of the parsed doc is our wrapper <root>
	root := doc.Root()
	if root == nil {
		return nil, ErrFailParseFragment
	}

	// Extract children of <root> — they are the fragment nodes
	var fragChildren []Node
	for child := root.FirstChild(); child != nil; {
		next := child.NextSibling() // save before detach clears Next
		// Detach from the parse document and re-parent to the target document
		childInner := getInternalNode(child)
		if childInner != nil {
			childInner.Detach()
			childInner.Doc = document.doc()
		}
		fragChildren = append(fragChildren, child)
		child = next
	}

	// Re-link siblings since Detach() broke the chain for detached nodes
	for i := 0; i < len(fragChildren)-1; i++ {
		inner := getInternalNode(fragChildren[i])
		nextInner := getInternalNode(fragChildren[i+1])
		if inner != nil && nextInner != nil {
			inner.Next = nextInner
			nextInner.Prev = inner
		}
	}

	// Create fragment with first child as root node
	if len(fragChildren) == 0 {
		return nil, ErrEmptyFragment
	}

	fragment = &DocumentFragment{
		Node:        fragChildren[0],
		InEncoding:  document.InputEncoding(),
		OutEncoding: document.OutputEncoding(),
	}

	document.BookkeepFragment(fragment)
	return
}

func ParseFragment(content, inEncoding, url []byte, options ParseOption, outEncoding []byte) (fragment *DocumentFragment, err error) {
	document := CreateEmptyDocument(inEncoding, outEncoding)
	fragment, err = parsefragment(document, nil, content, url, options)
	return
}

func (fragment *DocumentFragment) Remove() {
	fragment.Node.Remove()
}

func (fragment *DocumentFragment) Children() []Node {
	nodes := make([]Node, 0, initChildrenNumber)
	// The fragment nodes are siblings under the original wrapper root.
	// fragment.Node is the first one; iterate through its siblings.
	for child := fragment.Node; child != nil; child = child.NextSibling() {
		nodes = append(nodes, child)
	}
	return nodes
}

func (fragment *DocumentFragment) FirstChild() Node {
	return fragment.Node.FirstChild()
}

func (fragment *DocumentFragment) ToBuffer(outputBuffer []byte) []byte {
	var b []byte
	var size int
	for _, node := range fragment.Children() {
		if docType := node.MyDocument().DocType(); docType == XML_HTML_DOCUMENT_NODE {
			b, size = node.ToHtml(fragment.OutEncoding, nil)
		} else {
			b, size = node.ToXml(fragment.OutEncoding, nil)
		}
		outputBuffer = append(outputBuffer, b[:size]...)
	}
	return outputBuffer
}

func (fragment *DocumentFragment) String() string {
	b := fragment.ToBuffer(nil)
	if b == nil {
		return ""
	}
	return string(b)
}

// Search delegates to the fragment's Node.
func (fragment *DocumentFragment) Search(data interface{}) ([]Node, error) {
	return fragment.Node.Search(data)
}

// Clean whitespace bytes helper
var whitespaceBytes = []byte(" \t\r\n")

func trimWhitespaceBytes(b []byte) []byte {
	for len(b) > 0 {
		if bytes.IndexByte(whitespaceBytes, b[0]) >= 0 {
			b = b[1:]
		} else {
			break
		}
	}
	for len(b) > 0 {
		if bytes.IndexByte(whitespaceBytes, b[len(b)-1]) >= 0 {
			b = b[:len(b)-1]
		} else {
			break
		}
	}
	return b
}

func init() {
	_ = strings.TrimSpace // keep import
}
