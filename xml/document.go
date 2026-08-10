package xml

import (
	"errors"
	"os"

	"github.com/freemed/gokogiri/xpath"
)

// Document is the public interface for XML/HTML documents.
type Document interface {
	CreateElementNode(string) *ElementNode
	CreateCDataNode(string) *CDataNode
	CreateTextNode(string) *TextNode
	CreateCommentNode(string) *CommentNode
	CreatePINode(string, string) *ProcessingInstructionNode
	ParseFragment([]byte, []byte, ParseOption) (*DocumentFragment, error)

	DocPtr() interface{}
	DocType() NodeType
	DocRef() Document
	InputEncoding() []byte
	OutputEncoding() []byte
	DocXPathCtx() *xpath.XPath
	AddUnlinkedNode(interface{})
	RemoveUnlinkedNode(interface{}) bool
	Free()
	String() string
	Root() *ElementNode
	NodeById(string) *ElementNode
	BookkeepFragment(*DocumentFragment)

	RecursivelyRemoveNamespaces() error
	UnparsedEntityURI(string) string
	Uri() string

	// internal accessor
	doc() *InternalDoc
}

// ParseOption values allow you to tune the behaviour of the parsing engine.
type ParseOption int

const (
	XML_PARSE_RECOVER    ParseOption = 1 << iota // recover on errors
	XML_PARSE_NOENT                              // substitute entities
	XML_PARSE_DTDLOAD                            // load the external subset
	XML_PARSE_DTDATTR                            // default DTD attributes
	XML_PARSE_DTDVALID                           // validate with the DTD
	XML_PARSE_NOERROR                            // suppress error reports
	XML_PARSE_NOWARNING                          // suppress warning reports
	XML_PARSE_PEDANTIC                           // pedantic error reporting
	XML_PARSE_NOBLANKS                           // remove blank nodes
	XML_PARSE_SAX1                               // use the SAX1 interface internally
	XML_PARSE_XINCLUDE                           // Implement XInclude substitution
	XML_PARSE_NONET                              // Forbid network access
	XML_PARSE_NODICT                             // Do not reuse the context dictionary
	XML_PARSE_NSCLEAN                            // remove redundant namespaces declarations
	XML_PARSE_NOCDATA                            // merge CDATA as text nodes
	XML_PARSE_NOXINCNODE                         // do not generate XINCLUDE START/END nodes
	XML_PARSE_COMPACT                            // compact small text nodes; makes tree read-only
	XML_PARSE_OLD10                              // parse using XML-1.0 before update 5
	XML_PARSE_NOBASEFIX                          // do not fixup XINCLUDE xml:base uris
	XML_PARSE_HUGE                               // relax any hardcoded limit from the parser
	XML_PARSE_OLDSAX                             // parse using SAX2 interface before 2.7.0
	XML_PARSE_IGNORE_ENC                         // ignore internal document encoding hint
	XML_PARSE_BIG_LINES                          // Store big lines numbers in text PSVI field
)

// DefaultParseOption provides liberal parsing highly tolerant of invalid documents.
const DefaultParseOption ParseOption = XML_PARSE_RECOVER |
	XML_PARSE_NONET |
	XML_PARSE_NOERROR |
	XML_PARSE_NOWARNING

// StrictParseOption provides standard-compliant parsing.
const StrictParseOption ParseOption = XML_PARSE_NOENT |
	XML_PARSE_DTDLOAD |
	XML_PARSE_DTDATTR |
	XML_PARSE_NOCDATA

// DefaultEncoding is UTF-8.
const DefaultEncoding = "utf-8"

var ERR_FAILED_TO_PARSE_XML = errors.New("failed to parse xml input")

// XmlDocument is the primary interface for working with XML documents.
type XmlDocument struct {
	inner        *InternalDoc
	Me           Document
	Node         Node
	InEncoding   []byte
	OutEncoding  []byte
	XPathCtx     *xpath.XPath
	Type         NodeType
	InputLen     int
	UnlinkedNodeKeys []*InternalNode // track unlinked nodes for Free()
}

// DefaultEncodingBytes for convenient passing.
var DefaultEncodingBytes = []byte(DefaultEncoding)

const initialFragments = 2

// NewDocument wraps an InternalDoc into an XmlDocument.
func NewDocument(inner *InternalDoc, contentLen int, inEncoding, outEncoding []byte) (doc *XmlDocument) {
	doc = &XmlDocument{
		inner:       inner,
		InEncoding:  inEncoding,
		OutEncoding: outEncoding,
		InputLen:    contentLen,
		Type:        inner.DocType,
	}
	// Create the root XmlNode
	if inner.Root != nil {
		xmlNode := &XmlNode{
			inner:    inner.Root,
			Document: doc,
			valid:    true,
		}
		doc.Node = xmlNode
	} else {
		// Create a placeholder node for the document itself
		doc.Node = &XmlNode{
			inner:    &InternalNode{Typ: inner.DocType, Doc: inner, Valid: true},
			Document: doc,
			valid:    true,
		}
	}
	doc.XPathCtx = xpath.NewXPath(doc)
	doc.Me = doc
	return
}

// Parse creates an XmlDocument from some pre-existing content.
func Parse(content, inEncoding, url []byte, options ParseOption, outEncoding []byte) (doc *XmlDocument, err error) {
	contentLen := len(content)
	if contentLen > 0 {
		doc, err = parseXML(content, inEncoding, url, options, outEncoding)
	} else {
		doc = CreateEmptyDocument(inEncoding, outEncoding)
	}
	return
}

// ReadFile loads an XmlDocument from a filename.
func ReadFile(filename string, options ParseOption) (doc *XmlDocument, err error) {
	_, err = os.Stat(filename)
	if err != nil {
		return
	}
	content, err := os.ReadFile(filename)
	if err != nil {
		return
	}
	return parseXML(content, nil, []byte(filename), options, DefaultEncodingBytes)
}

// CreateEmptyDocument creates an empty XML document.
func CreateEmptyDocument(inEncoding, outEncoding []byte) (doc *XmlDocument) {
	inner := &InternalDoc{
		DocType:       XML_DOCUMENT_NODE,
		UnlinkedNodes: make(map[*InternalNode]bool),
	}
	return NewDocument(inner, 0, inEncoding, outEncoding)
}

// DocPtr provides access to the underlying document structure.
func (document *XmlDocument) DocPtr() (ptr interface{}) {
	ptr = document.inner
	return
}

// DocType returns the node type constant.
func (document *XmlDocument) DocType() (t NodeType) {
	t = document.Type
	return
}

// DocRef returns the embedded Document interface.
func (document *XmlDocument) DocRef() (d Document) {
	d = document.Me
	return
}

// InputEncoding is the original encoding of the document.
func (document *XmlDocument) InputEncoding() (encoding []byte) {
	encoding = document.InEncoding
	return
}

// OutputEncoding is the encoding that will be used when serializing.
func (document *XmlDocument) OutputEncoding() (encoding []byte) {
	encoding = document.OutEncoding
	return
}

// DocXPathCtx returns the XPath context.
func (document *XmlDocument) DocXPathCtx() (ctx *xpath.XPath) {
	ctx = document.XPathCtx
	return
}

func (document *XmlDocument) AddUnlinkedNode(nodePtr interface{}) {
	if inner, ok := nodePtr.(*InternalNode); ok {
		document.UnlinkedNodeKeys = append(document.UnlinkedNodeKeys, inner)
	}
}

func (document *XmlDocument) RemoveUnlinkedNode(nodePtr interface{}) bool {
	if inner, ok := nodePtr.(*InternalNode); ok {
		for i, n := range document.UnlinkedNodeKeys {
			if n == inner {
				document.UnlinkedNodeKeys = append(document.UnlinkedNodeKeys[:i], document.UnlinkedNodeKeys[i+1:]...)
				return true
			}
		}
	}
	return false
}

func (document *XmlDocument) BookkeepFragment(fragment *DocumentFragment) {
	document.inner.Fragments = append(document.inner.Fragments, fragment)
}

// Root returns the root element of the document.
func (document *XmlDocument) Root() (element *ElementNode) {
	if document.inner != nil && document.inner.Root != nil {
		elem := NewNode(document.inner.Root, document)
		if e, ok := elem.(*ElementNode); ok {
			element = e
		}
	}
	return
}

// NodeById finds an element by its ID attribute value.
func (document *XmlDocument) NodeById(id string) (element *ElementNode) {
	if document.inner == nil || document.inner.IDIndex == nil {
		return nil
	}
	if n, found := document.inner.IDIndex[id]; found {
		elem := NewNode(n, document)
		if e, ok := elem.(*ElementNode); ok {
			element = e
		}
	}
	return
}

// CreateElementNode creates an element node.
func (document *XmlDocument) CreateElementNode(tag string) (element *ElementNode) {
	inner := &InternalNode{
		Typ:   XML_ELEMENT_NODE,
		Name:  tag,
		Doc: document.inner,
		Valid: true,
	}
	node := NewNode(inner, document)
	element = node.(*ElementNode)
	return
}

// CreateTextNode creates a text node.
func (document *XmlDocument) CreateTextNode(data string) (text *TextNode) {
	inner := &InternalNode{
		Typ:     XML_TEXT_NODE,
		Content: data,
		Doc: document.inner,
		Valid:   true,
	}
	node := NewNode(inner, document)
	text = node.(*TextNode)
	return
}

// CreateCDataNode creates a CDATA node.
func (document *XmlDocument) CreateCDataNode(data string) (cdata *CDataNode) {
	inner := &InternalNode{
		Typ:     XML_CDATA_SECTION_NODE,
		Content: data,
		Doc: document.inner,
		Valid:   true,
	}
	node := NewNode(inner, document)
	cdata = node.(*CDataNode)
	return
}

// CreateCommentNode creates a comment node.
func (document *XmlDocument) CreateCommentNode(data string) (comment *CommentNode) {
	inner := &InternalNode{
		Typ:     XML_COMMENT_NODE,
		Content: data,
		Doc: document.inner,
		Valid:   true,
	}
	node := NewNode(inner, document)
	comment = node.(*CommentNode)
	return
}

// CreatePINode creates a processing instruction node.
func (document *XmlDocument) CreatePINode(name, data string) (pi *ProcessingInstructionNode) {
	inner := &InternalNode{
		Typ:     XML_PI_NODE,
		Name:    name,
		Content: data,
		Doc: document.inner,
		Valid:   true,
	}
	node := NewNode(inner, document)
	pi = node.(*ProcessingInstructionNode)
	return
}

func (document *XmlDocument) ParseFragment(input, url []byte, options ParseOption) (fragment *DocumentFragment, err error) {
	root := document.Root()
	if root == nil {
		fragment, err = parsefragment(document, nil, input, url, options)
	} else {
		fragment, err = parsefragment(document, root.XmlNode, input, url, options)
	}
	return
}

// UnparsedEntityURI returns the URI of an NDATA entity from the DTD.
func (document *XmlDocument) UnparsedEntityURI(name string) (val string) {
	if document.inner == nil || document.inner.DTDInfo == nil || document.inner.DTDInfo.Entities == nil {
		return ""
	}
	if entity, ok := document.inner.DTDInfo.Entities[name]; ok && entity.IsNDATA {
		return entity.SystemID
	}
	return ""
}

// Free releases resources associated with the document.
func (document *XmlDocument) Free() {
	if document.XPathCtx != nil {
		document.XPathCtx.Free()
		document.XPathCtx = nil
	}
	if document.inner != nil {
		// Free fragments
		for _, fragment := range document.inner.Fragments {
			fragment.Remove()
		}
		document.inner.Fragments = nil
		// Free unlinked nodes (already handled by Go GC, but clear refs)
		document.UnlinkedNodeKeys = nil
		document.inner = nil
	}
}

// Uri returns the document URI.
func (document *XmlDocument) Uri() (val string) {
	if document.inner != nil {
		val = document.inner.Url
	}
	return
}

// doc returns the internal document pointer.
func (document *XmlDocument) doc() *InternalDoc {
	return document.inner
}

// RecursivelyRemoveNamespaces delegates to the root element.
func (document *XmlDocument) RecursivelyRemoveNamespaces() (err error) {
	root := document.Root()
	if root != nil {
		return root.RecursivelyRemoveNamespaces()
	}
	return nil
}

// String delegates to the embedded Node.
func (document *XmlDocument) String() string {
	if document.Node != nil {
		return document.Node.String()
	}
	return ""
}

// ToBuffer delegates to the embedded Node.
func (document *XmlDocument) ToBuffer(outputBuffer []byte) []byte {
	if document.Node != nil {
		return document.Node.ToBuffer(outputBuffer)
	}
	return nil
}

// Search delegates to the embedded Node.
func (document *XmlDocument) Search(data interface{}) ([]Node, error) {
	if document.Node != nil {
		return document.Node.Search(data)
	}
	return nil, nil
}

// SearchWithVariables delegates to the embedded Node.
func (document *XmlDocument) SearchWithVariables(data interface{}, v interface{}) ([]Node, error) {
	if document.Node != nil {
		return document.Node.SearchWithVariables(data, nil)
	}
	return nil, nil
}

// EvalXPath delegates to the embedded Node.
func (document *XmlDocument) EvalXPath(data interface{}, v interface{}) (interface{}, error) {
	if document.Node != nil {
		return document.Node.EvalXPath(data, nil)
	}
	return nil, nil
}

// EvalXPathAsBoolean delegates to the embedded Node.
func (document *XmlDocument) EvalXPathAsBoolean(data interface{}, v interface{}) bool {
	if document.Node != nil {
		return document.Node.EvalXPathAsBoolean(data, nil)
	}
	return false
}
