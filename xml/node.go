package xml

import (
	"errors"
	"strconv"
	"sync"
	"unicode/utf8"

	"github.com/freemed/gokogiri/xpath"
)

// NodeType is an enumeration that indicates the type of XmlNode.
type NodeType int

const (
	XML_ELEMENT_NODE       NodeType = iota + 1
	XML_ATTRIBUTE_NODE
	XML_TEXT_NODE
	XML_CDATA_SECTION_NODE
	XML_ENTITY_REF_NODE
	XML_ENTITY_NODE
	XML_PI_NODE
	XML_COMMENT_NODE
	XML_DOCUMENT_NODE
	XML_DOCUMENT_TYPE_NODE
	XML_DOCUMENT_FRAG_NODE
	XML_NOTATION_NODE
	XML_HTML_DOCUMENT_NODE
	XML_DTD_NODE
	XML_ELEMENT_DECL
	XML_ATTRIBUTE_DECL
	XML_ENTITY_DECL
	XML_NAMESPACE_DECL
	XML_XINCLUDE_START
	XML_XINCLUDE_END
	XML_DOCB_DOCUMENT_NODE
)

// SerializationOption is a set of flags used to control how a node is written out.
type SerializationOption int

const (
	XML_SAVE_FORMAT   SerializationOption = 1 << iota // format save output
	XML_SAVE_NO_DECL                                  // drop the xml declaration
	XML_SAVE_NO_EMPTY                                 // no empty tags
	XML_SAVE_NO_XHTML                                 // disable XHTML1 specific rules
	XML_SAVE_XHTML                                    // force XHTML1 specific rules
	XML_SAVE_AS_XML                                   // force XML serialization on HTML doc
	XML_SAVE_AS_HTML                                  // force HTML serialization on XML doc
	XML_SAVE_WSNONSIG                                 // format with non-significant whitespace
)

// NamespaceDeclaration represents a namespace declaration.
type NamespaceDeclaration struct {
	Prefix string
	Uri    string
}

// Node is the public interface for all node types.
type Node interface {
	NodePtr() interface{}
	ResetNodePtr()
	MyDocument() Document

	IsValid() bool

	ParseFragment([]byte, []byte, ParseOption) (*DocumentFragment, error)
	LineNumber() int

	NodeType() NodeType
	NextSibling() Node
	PreviousSibling() Node

	Parent() Node
	FirstChild() Node
	LastChild() Node
	CountChildren() int
	Attributes() map[string]*AttributeNode
	AttributeList() []*AttributeNode

	Coerce(interface{}) ([]Node, error)

	AddChild(interface{}) error
	AddPreviousSibling(interface{}) error
	AddNextSibling(interface{}) error
	InsertBefore(interface{}) error
	InsertAfter(interface{}) error
	InsertBegin(interface{}) error
	InsertEnd(interface{}) error
	SetInnerHtml(interface{}) error
	SetChildren(interface{}) error
	Replace(interface{}) error
	Wrap(string) error

	SetContent(interface{}) error

	Name() string
	SetName(string)

	Attr(string) string
	SetAttr(string, string) string
	SetNsAttr(string, string, string) string
	Attribute(string) *AttributeNode

	Path() string

	Duplicate(int) Node
	DuplicateTo(Document, int) Node

	Search(interface{}) ([]Node, error)
	SearchWithVariables(interface{}, xpath.VariableScope) ([]Node, error)
	EvalXPath(interface{}, xpath.VariableScope) (interface{}, error)
	EvalXPathAsBoolean(interface{}, xpath.VariableScope) bool

	Unlink()
	Remove()
	ResetChildren()

	SerializeWithFormat(SerializationOption, []byte, []byte) ([]byte, int)
	ToXml([]byte, []byte) ([]byte, int)
	ToUnformattedXml() string
	ToHtml([]byte, []byte) ([]byte, int)
	ToBuffer([]byte) []byte
	String() string
	Content() string
	InnerHtml() string

	RecursivelyRemoveNamespaces() error
	Namespace() string
	SetNamespace(string, string)
	DeclareNamespace(string, string)
	RemoveDefaultNamespace()
	DeclaredNamespaces() []NamespaceDeclaration
}

var (
	ERR_UNDEFINED_COERCE_PARAM               = errors.New("unexpected parameter type in coerce")
	ERR_UNDEFINED_SET_CONTENT_PARAM          = errors.New("unexpected parameter type in SetContent")
	ERR_UNDEFINED_SEARCH_PARAM               = errors.New("unexpected parameter type in Search")
	ERR_CANNOT_MAKE_DUCMENT_AS_CHILD         = errors.New("cannot add a document node as a child")
	ERR_CANNOT_COPY_TEXT_NODE_WHEN_ADD_CHILD = errors.New("cannot copy a text node when adding it")
)

// run out of memory
var ErrTooLarge = errors.New("Output buffer too large")

/*
XmlNode implements the Node interface, and as such is the heart of the API.
*/
type XmlNode struct {
	inner    *InternalNode
	Document Document
	valid    bool
}

// WriteBuffer for serialization callbacks.
type WriteBuffer struct {
	Node   *XmlNode
	Buffer []byte
	Offset int
}

var (
	wbuffer      *WriteBuffer
	wbufferMutex sync.Mutex
)

// NewNode creates a Node from an InternalNode pointer.
func NewNode(inner *InternalNode, document Document) (node Node) {
	if inner == nil {
		return nil
	}
	xmlNode := &XmlNode{
		inner:    inner,
		Document: document,
		valid:    inner.Valid,
	}
	nodeType := inner.Typ

	switch nodeType {
	default:
		node = xmlNode
	case XML_ATTRIBUTE_NODE:
		node = &AttributeNode{XmlNode: xmlNode}
	case XML_ELEMENT_NODE:
		node = &ElementNode{XmlNode: xmlNode}
	case XML_CDATA_SECTION_NODE:
		node = &CDataNode{XmlNode: xmlNode}
	case XML_COMMENT_NODE:
		node = &CommentNode{XmlNode: xmlNode}
	case XML_PI_NODE:
		node = &ProcessingInstructionNode{XmlNode: xmlNode}
	case XML_TEXT_NODE:
		node = &TextNode{XmlNode: xmlNode}
	}
	return
}

// newNodeFromInternal is a convenience for wrapping an InternalNode into an XmlNode (not typed).
func newNodeFromInternal(inner *InternalNode, doc Document) *XmlNode {
	if inner == nil {
		return nil
	}
	return &XmlNode{
		inner:    inner,
		Document: doc,
		valid:    inner.Valid,
	}
}

func (xmlNode *XmlNode) coerce(data interface{}) (nodes []Node, err error) {
	switch t := data.(type) {
	default:
		err = ERR_UNDEFINED_COERCE_PARAM
	case []Node:
		nodes = t
	case *DocumentFragment:
		nodes = t.Children()
	case string:
		f, err := xmlNode.MyDocument().ParseFragment([]byte(t), nil, DefaultParseOption)
		if err == nil {
			nodes = f.Children()
		}
	case []byte:
		f, err := xmlNode.MyDocument().ParseFragment(t, nil, DefaultParseOption)
		if err == nil {
			nodes = f.Children()
		}
	}
	return
}

func (xmlNode *XmlNode) Coerce(data interface{}) (nodes []Node, err error) {
	return xmlNode.coerce(data)
}

func (xmlNode *XmlNode) AddChild(data interface{}) (err error) {
	switch t := data.(type) {
	default:
		if nodes, cerr := xmlNode.coerce(data); cerr == nil {
			for _, node := range nodes {
				if err = xmlNode.addChild(node); err != nil {
					break
				}
			}
		} else {
			err = cerr
		}
	case *DocumentFragment:
		if nodes, cerr := xmlNode.coerce(data); cerr == nil {
			for _, node := range nodes {
				if err = xmlNode.addChild(node); err != nil {
					break
				}
			}
		} else {
			err = cerr
		}
	case Node:
		err = xmlNode.addChild(t)
	}
	return
}

func (xmlNode *XmlNode) AddPreviousSibling(data interface{}) (err error) {
	switch t := data.(type) {
	default:
		if nodes, cerr := xmlNode.coerce(data); cerr == nil {
			for _, node := range nodes {
				if err = xmlNode.addPreviousSibling(node); err != nil {
					break
				}
			}
		} else {
			err = cerr
		}
	case *DocumentFragment:
		if nodes, cerr := xmlNode.coerce(data); cerr == nil {
			for _, node := range nodes {
				if err = xmlNode.addPreviousSibling(node); err != nil {
					break
				}
			}
		} else {
			err = cerr
		}
	case Node:
		err = xmlNode.addPreviousSibling(t)
	}
	return
}

func (xmlNode *XmlNode) AddNextSibling(data interface{}) (err error) {
	switch t := data.(type) {
	default:
		if nodes, cerr := xmlNode.coerce(data); cerr == nil {
			for i := len(nodes) - 1; i >= 0; i-- {
				node := nodes[i]
				if err = xmlNode.addNextSibling(node); err != nil {
					break
				}
			}
		} else {
			err = cerr
		}
	case *DocumentFragment:
		if nodes, cerr := xmlNode.coerce(data); cerr == nil {
			for i := len(nodes) - 1; i >= 0; i-- {
				node := nodes[i]
				if err = xmlNode.addNextSibling(node); err != nil {
					break
				}
			}
		} else {
			err = cerr
		}
	case Node:
		err = xmlNode.addNextSibling(t)
	}
	return
}

func (xmlNode *XmlNode) ResetNodePtr() {
	xmlNode.inner = nil
}

func (xmlNode *XmlNode) IsValid() bool {
	return xmlNode.valid && xmlNode.inner != nil && xmlNode.inner.Valid
}

func (xmlNode *XmlNode) MyDocument() (document Document) {
	document = xmlNode.Document.DocRef()
	return
}

// NodePtr returns the internal node pointer (was unsafe.Pointer to C struct).
func (xmlNode *XmlNode) NodePtr() (p interface{}) {
	p = xmlNode.inner
	return
}

func (xmlNode *XmlNode) NodeType() (nodeType NodeType) {
	if xmlNode.inner == nil {
		return 0
	}
	nodeType = xmlNode.inner.Typ
	return
}

// Path returns an XPath expression that selects this node.
func (xmlNode *XmlNode) Path() (path string) {
	if xmlNode.inner == nil {
		return ""
	}
	return buildNodePath(xmlNode.inner)
}

// buildNodePath generates an XPath like /root/child[1]/target.
func buildNodePath(n *InternalNode) string {
	if n.Parent == nil {
		if n.Doc != nil && n.Doc.Root != nil {
			return "/" + n.Doc.Root.Name
		}
		return ""
	}
	parentPath := buildNodePath(n.Parent)
	// Count preceding siblings of same name
	idx := 1
	for s := n.Prev; s != nil; s = s.Prev {
		if s.Typ == XML_ELEMENT_NODE && s.Name == n.Name {
			idx++
		}
	}
	return parentPath + "/" + n.Name + "[" + strconv.Itoa(idx) + "]"
}

func (xmlNode *XmlNode) NextSibling() Node {
	if xmlNode.inner == nil {
		return nil
	}
	return NewNode(xmlNode.inner.Next, xmlNode.Document)
}

func (xmlNode *XmlNode) PreviousSibling() Node {
	if xmlNode.inner == nil {
		return nil
	}
	return NewNode(xmlNode.inner.Prev, xmlNode.Document)
}

func (xmlNode *XmlNode) CountChildren() int {
	if xmlNode.inner == nil {
		return 0
	}
	return xmlNode.inner.CountChildElements()
}

func (xmlNode *XmlNode) FirstChild() Node {
	if xmlNode.inner == nil {
		return nil
	}
	return NewNode(xmlNode.inner.Children, xmlNode.Document)
}

func (xmlNode *XmlNode) LastChild() Node {
	if xmlNode.inner == nil {
		return nil
	}
	return NewNode(xmlNode.inner.Last, xmlNode.Document)
}

func (xmlNode *XmlNode) Parent() Node {
	if xmlNode.inner == nil || xmlNode.inner.Parent == nil {
		return nil
	}
	// Check for sentinel parent value (-1 pointer in original C)
	return NewNode(xmlNode.inner.Parent, xmlNode.Document)
}

func (xmlNode *XmlNode) ResetChildren() {
	if xmlNode.inner == nil {
		return
	}
	children := xmlNode.inner.RemoveChildren()
	doc := xmlNode.Document
	for _, child := range children {
		doc.AddUnlinkedNode(child)
	}
}

var (
	contentNode  *XmlNode
	contentMutex sync.Mutex
)

func (xmlNode *XmlNode) SetContent(content interface{}) (err error) {
	switch data := content.(type) {
	default:
		err = ERR_UNDEFINED_SET_CONTENT_PARAM
	case string:
		xmlNode.ResetChildren()
		if xmlNode.inner != nil {
			textNode := &InternalNode{
				Typ:     XML_TEXT_NODE,
				Content: data,
				Doc: xmlNode.inner.Doc,
				Valid:   true,
			}
			xmlNode.inner.AppendChild(textNode)
		}
	case []byte:
		err = xmlNode.SetContent(string(data))
	}
	return
}

func (xmlNode *XmlNode) InsertBefore(data interface{}) (err error) {
	return xmlNode.AddPreviousSibling(data)
}

func (xmlNode *XmlNode) InsertAfter(data interface{}) (err error) {
	return xmlNode.AddNextSibling(data)
}

func (xmlNode *XmlNode) InsertBegin(data interface{}) (err error) {
	if parent := xmlNode.Parent(); parent != nil {
		if last := parent.LastChild(); last != nil {
			err = last.AddPreviousSibling(data)
		}
	}
	return
}

func (xmlNode *XmlNode) InsertEnd(data interface{}) (err error) {
	if parent := xmlNode.Parent(); parent != nil {
		if first := parent.FirstChild(); first != nil {
			err = first.AddPreviousSibling(data)
		}
	}
	return
}

func (xmlNode *XmlNode) SetChildren(data interface{}) (err error) {
	nodes, err := xmlNode.coerce(data)
	if err != nil {
		return
	}
	xmlNode.ResetChildren()
	err = xmlNode.AddChild(nodes)
	return nil
}

func (xmlNode *XmlNode) SetInnerHtml(data interface{}) (err error) {
	return xmlNode.SetChildren(data)
}

func (xmlNode *XmlNode) Replace(data interface{}) (err error) {
	err = xmlNode.AddPreviousSibling(data)
	if err != nil {
		return
	}
	xmlNode.Remove()
	return
}

// AttributeList returns document-ordered list of attribute nodes.
func (xmlNode *XmlNode) AttributeList() (attributes []*AttributeNode) {
	if xmlNode.inner == nil {
		return
	}
	for i := range xmlNode.inner.Props {
		prop := xmlNode.inner.Props[i]
		attrInner := &InternalNode{
			Typ:     XML_ATTRIBUTE_NODE,
			Name:    prop.Name,
			Content: prop.Value,
			Doc:     xmlNode.inner.Doc,
			Valid:   true,
		}
		if prop.Ns != nil {
			attrInner.Ns = &InternalNs{Prefix: prop.Ns.Prefix, Href: prop.Ns.Href}
		}
		attrNode := &AttributeNode{
			XmlNode: &XmlNode{
				inner:    attrInner,
				Document: xmlNode.Document,
				valid:    true,
			},
			sourceAttr: prop, // Link back to source for write-through
		}
		attributes = append(attributes, attrNode)
	}
	return
}

// Attributes returns attribute nodes indexed by name.
func (xmlNode *XmlNode) Attributes() (attributes map[string]*AttributeNode) {
	attributes = make(map[string]*AttributeNode)
	if xmlNode.inner == nil {
		return
	}
	for i := range xmlNode.inner.Props {
		prop := xmlNode.inner.Props[i]
		attrInner := &InternalNode{
			Typ:     XML_ATTRIBUTE_NODE,
			Name:    prop.Name,
			Content: prop.Value,
			Doc:     xmlNode.inner.Doc,
			Valid:   true,
		}
		if prop.Ns != nil {
			attrInner.Ns = &InternalNs{Prefix: prop.Ns.Prefix, Href: prop.Ns.Href}
		}
		attrNode := &AttributeNode{
			XmlNode: &XmlNode{
				inner:    attrInner,
				Document: xmlNode.Document,
				valid:    true,
			},
			sourceAttr: prop,
		}
		attributes[prop.Name] = attrNode
	}
	return
}

// Attribute returns a specific attribute node, or nil.
func (xmlNode *XmlNode) Attribute(name string) (attribute *AttributeNode) {
	if xmlNode.inner == nil || xmlNode.inner.Typ != XML_ELEMENT_NODE {
		return
	}
	val, found := xmlNode.inner.GetAttr(name)
	if !found {
		return nil
	}
	attrInner := &InternalNode{
		Typ:     XML_ATTRIBUTE_NODE,
		Name:    name,
		Content: val,
		Doc: xmlNode.inner.Doc,
		Valid:   true,
	}
	node := NewNode(attrInner, xmlNode.Document)
	if n, ok := node.(*AttributeNode); ok {
		attribute = n
	}
	return
}

// Attr returns the value of an attribute.
func (xmlNode *XmlNode) Attr(name string) (val string) {
	if xmlNode.inner == nil || xmlNode.inner.Typ != XML_ELEMENT_NODE {
		return
	}
	val, _ = xmlNode.inner.GetAttr(name)
	return
}

// SetAttr sets an attribute value.
func (xmlNode *XmlNode) SetAttr(name, value string) (val string) {
	val = value
	if xmlNode.inner == nil || xmlNode.inner.Typ != XML_ELEMENT_NODE {
		return
	}
	xmlNode.inner.SetAttr(name, value)
	return
}

// SetNsAttr sets a namespaced attribute.
func (xmlNode *XmlNode) SetNsAttr(href, name, value string) (val string) {
	val = value
	if xmlNode.inner == nil || xmlNode.inner.Typ != XML_ELEMENT_NODE {
		return
	}
	// Ensure namespace declaration exists
	ns := xmlNode.inner.LookupNamespaceByHref(href)
	if ns == nil {
		return
	}
	xmlNode.inner.SetNsAttr(href, name, value)
	return
}

// Search for nodes matching an XPath.
func (xmlNode *XmlNode) Search(data interface{}) (result []Node, err error) {
	switch data := data.(type) {
	default:
		err = ERR_UNDEFINED_SEARCH_PARAM
	case string:
		if xpathExpr := xpath.Compile(data); xpathExpr != nil {
			defer xpathExpr.Free()
			result, err = xmlNode.Search(xpathExpr)
		} else {
			err = errors.New("cannot compile xpath: " + data)
		}
	case []byte:
		result, err = xmlNode.Search(string(data))
	case *xpath.Expression:
		xpathCtx := xmlNode.Document.DocXPathCtx()
		if xpathCtx == nil {
			xpathCtx = xpath.NewXPath(xmlNode.Document)
		}
		nodePtrs, xerr := xpathCtx.EvaluateAsNodeset(xmlNode.inner, data)
		if nodePtrs == nil || xerr != nil {
			return nil, xerr
		}
		for _, nodePtr := range nodePtrs {
			if in, ok := nodePtr.(*InternalNode); ok {
				result = append(result, NewNode(in, xmlNode.Document))
			}
		}
	}
	return
}

// SearchWithVariables searches with variable scope.
func (xmlNode *XmlNode) SearchWithVariables(data interface{}, v xpath.VariableScope) (result []Node, err error) {
	switch data := data.(type) {
	default:
		err = ERR_UNDEFINED_SEARCH_PARAM
	case string:
		if xpathExpr := xpath.Compile(data); xpathExpr != nil {
			defer xpathExpr.Free()
			result, err = xmlNode.SearchWithVariables(xpathExpr, v)
		} else {
			err = errors.New("cannot compile xpath: " + data)
		}
	case []byte:
		result, err = xmlNode.SearchWithVariables(string(data), v)
	case *xpath.Expression:
		xpathCtx := xmlNode.Document.DocXPathCtx()
		if xpathCtx == nil {
			xpathCtx = xpath.NewXPath(xmlNode.Document)
		}
		xpathCtx.SetResolver(v)
		nodePtrs, xerr := xpathCtx.EvaluateAsNodeset(xmlNode.inner, data)
		if nodePtrs == nil || xerr != nil {
			return nil, xerr
		}
		for _, nodePtr := range nodePtrs {
			if in, ok := nodePtr.(*InternalNode); ok {
				result = append(result, NewNode(in, xmlNode.Document))
			}
		}
	}
	return
}

// EvalXPath evaluates an XPath and returns a typed result.
func (xmlNode *XmlNode) EvalXPath(data interface{}, v xpath.VariableScope) (result interface{}, err error) {
	switch data := data.(type) {
	case string:
		if xpathExpr := xpath.Compile(data); xpathExpr != nil {
			defer xpathExpr.Free()
			result, err = xmlNode.EvalXPath(xpathExpr, v)
		} else {
			err = errors.New("cannot compile xpath: " + data)
		}
	case []byte:
		result, err = xmlNode.EvalXPath(string(data), v)
	case *xpath.Expression:
		xpathCtx := xmlNode.Document.DocXPathCtx()
		if xpathCtx == nil {
			xpathCtx = xpath.NewXPath(xmlNode.Document)
		}
		xpathCtx.SetResolver(v)
		xerr := xpathCtx.Evaluate(xmlNode.inner, data)
		if xerr != nil {
			return nil, xerr
		}
		rt := xpathCtx.ReturnType()
		switch rt {
		case xpath.XPATH_NODESET, xpath.XPATH_XSLT_TREE:
			nodePtrs, nerr := xpathCtx.ResultAsNodeset()
			if nerr != nil {
				return nil, nerr
			}
			var output []Node
			for _, nodePtr := range nodePtrs {
				if in, ok := nodePtr.(*InternalNode); ok {
					output = append(output, NewNode(in, xmlNode.Document))
				}
			}
			result = output
		case xpath.XPATH_NUMBER:
			result, _ = xpathCtx.ResultAsNumber()
		case xpath.XPATH_BOOLEAN:
			result, _ = xpathCtx.ResultAsBoolean()
		default:
			result, _ = xpathCtx.ResultAsString()
		}
	default:
		err = ERR_UNDEFINED_SEARCH_PARAM
	}
	return
}

// EvalXPathAsBoolean evaluates an XPath and coerces to boolean.
func (xmlNode *XmlNode) EvalXPathAsBoolean(data interface{}, v xpath.VariableScope) (result bool) {
	switch data := data.(type) {
	case string:
		if xpathExpr := xpath.Compile(data); xpathExpr != nil {
			defer xpathExpr.Free()
			result = xmlNode.EvalXPathAsBoolean(xpathExpr, v)
		}
	case []byte:
		result = xmlNode.EvalXPathAsBoolean(string(data), v)
	case *xpath.Expression:
		xpathCtx := xmlNode.Document.DocXPathCtx()
		if xpathCtx == nil {
			xpathCtx = xpath.NewXPath(xmlNode.Document)
		}
		xpathCtx.SetResolver(v)
		err := xpathCtx.Evaluate(xmlNode.inner, data)
		if err != nil {
			return false
		}
		result, _ = xpathCtx.ResultAsBoolean()
	}
	return
}

// Name returns the local name of the node.
func (xmlNode *XmlNode) Name() (name string) {
	if xmlNode.inner != nil {
		name = xmlNode.inner.Name
	}
	return
}

// Namespace returns the namespace href.
func (xmlNode *XmlNode) Namespace() (href string) {
	if xmlNode.inner != nil && xmlNode.inner.Ns != nil {
		href = xmlNode.inner.Ns.Href
	}
	return
}

// SetName sets the local name.
func (xmlNode *XmlNode) SetName(name string) {
	if xmlNode.inner != nil {
		xmlNode.inner.Name = name
	}
}

func (xmlNode *XmlNode) Duplicate(level int) Node {
	return xmlNode.DuplicateTo(xmlNode.Document, level)
}

func (xmlNode *XmlNode) DuplicateTo(doc Document, level int) (dup Node) {
	if xmlNode.inner == nil || !xmlNode.inner.Valid {
		return nil
	}
	targetDoc := doc.doc()
	if targetDoc == nil {
		return nil
	}
	cp := xmlNode.inner.DeepCopy(targetDoc, level)
	return NewNode(cp, doc)
}

// serialize a node to a byte buffer.
func (xmlNode *XmlNode) serialize(format SerializationOption, encoding, outputBuffer []byte) ([]byte, int) {
	if xmlNode.inner == nil {
		return nil, 0
	}
	if len(encoding) == 0 {
		docEncoding := xmlNode.Document.OutputEncoding()
		if len(docEncoding) > 0 {
			encoding = docEncoding
		}
	}

	wbufferMutex.Lock()
	defer wbufferMutex.Unlock()
	if outputBuffer == nil {
		outputBuffer = make([]byte, 0)
	}
	wbuffer = &WriteBuffer{Node: xmlNode, Buffer: outputBuffer}

	// Delegate to the serialization engine
	var out []byte
	if format&XML_SAVE_AS_HTML != 0 {
		out = serializeToHTML(xmlNode.inner, format, encoding)
	} else {
		out = serializeToXML(xmlNode.inner, format, encoding)
	}

	wbuffer = nil
	return out, len(out)
}

func (xmlNode *XmlNode) SerializeWithFormat(format SerializationOption, encoding, outputBuffer []byte) ([]byte, int) {
	return xmlNode.serialize(format, encoding, outputBuffer)
}

func (xmlNode *XmlNode) ToXml(encoding, outputBuffer []byte) ([]byte, int) {
	// Don't include XML declaration when serializing individual nodes.
	// The declaration is added at the document level by XmlDocument.String().
	return xmlNode.serialize(XML_SAVE_AS_XML|XML_SAVE_FORMAT|XML_SAVE_NO_DECL, encoding, outputBuffer)
}

func (xmlNode *XmlNode) ToUnformattedXml() string {
	b, size := xmlNode.serialize(XML_SAVE_AS_XML|XML_SAVE_NO_DECL, nil, nil)
	if b == nil {
		return ""
	}
	return string(b[:size])
}

func (xmlNode *XmlNode) ToHtml(encoding, outputBuffer []byte) ([]byte, int) {
	return xmlNode.serialize(XML_SAVE_AS_HTML|XML_SAVE_FORMAT, encoding, outputBuffer)
}

func (xmlNode *XmlNode) ToBuffer(outputBuffer []byte) []byte {
	var b []byte
	var size int
	if docType := xmlNode.Document.DocType(); docType == XML_HTML_DOCUMENT_NODE {
		b, size = xmlNode.ToHtml(nil, outputBuffer)
	} else {
		b, size = xmlNode.ToXml(nil, outputBuffer)
	}
	return b[:size]
}

func (xmlNode *XmlNode) String() string {
	b := xmlNode.ToBuffer(nil)
	if b == nil {
		return ""
	}
	return string(b)
}

func (xmlNode *XmlNode) Content() string {
	if xmlNode.inner == nil {
		return ""
	}
	// For element nodes, collect text from all descendant text nodes
	if xmlNode.inner.Typ == XML_ELEMENT_NODE || xmlNode.inner.Typ == XML_DOCUMENT_NODE {
		return collectTextContent(xmlNode.inner)
	}
	return xmlNode.inner.Content
}

// collectTextContent recursively collects all text from descendant text/cdata nodes.
func collectTextContent(n *InternalNode) string {
	if n == nil {
		return ""
	}
	var buf string
	for c := n.Children; c != nil; c = c.Next {
		if c.Typ == XML_TEXT_NODE || c.Typ == XML_CDATA_SECTION_NODE {
			buf += c.Content
		} else if c.Typ == XML_ELEMENT_NODE {
			buf += collectTextContent(c)
		}
	}
	return buf
}

func (xmlNode *XmlNode) InnerHtml() string {
	out := ""
	for child := xmlNode.FirstChild(); child != nil; child = child.NextSibling() {
		out += child.String()
	}
	return out
}

func (xmlNode *XmlNode) Unlink() {
	if xmlNode.inner == nil {
		return
	}
	xmlNode.inner.Detach()
	xmlNode.Document.AddUnlinkedNode(xmlNode.inner)
}

func (xmlNode *XmlNode) Remove() {
	if xmlNode.inner == nil || !xmlNode.valid {
		return
	}
	docPtr := xmlNode.Document.doc()
	if docPtr != nil && xmlNode.inner == docPtr.Root {
		return // don't remove the document root
	}
	xmlNode.Unlink()
	xmlNode.valid = false
	xmlNode.inner.Valid = false
}

func (xmlNode *XmlNode) addChild(node Node) (err error) {
	nodeType := node.NodeType()
	if nodeType == XML_DOCUMENT_NODE || nodeType == XML_HTML_DOCUMENT_NODE {
		err = ERR_CANNOT_MAKE_DUCMENT_AS_CHILD
		return
	}
	childInner := getInternalNode(node)
	if childInner == nil || xmlNode.inner == nil {
		return
	}
	if xmlNode.inner == childInner {
		return
	}
	// Check if node is an ancestor
	if childInner.IsAncestor(xmlNode.inner) {
		node.Remove()
		return
	}
	// Remove from unlinked nodes if tracked
	xmlNode.Document.RemoveUnlinkedNode(childInner)
	xmlNode.inner.AppendChild(childInner)
	childInner.Valid = true
	return
}

func (xmlNode *XmlNode) addPreviousSibling(node Node) (err error) {
	nodeType := node.NodeType()
	if nodeType == XML_DOCUMENT_NODE || nodeType == XML_HTML_DOCUMENT_NODE {
		err = ERR_CANNOT_MAKE_DUCMENT_AS_CHILD
		return
	}
	siblingInner := getInternalNode(node)
	if siblingInner == nil || xmlNode.inner == nil || xmlNode.inner.Parent == nil {
		return
	}
	if xmlNode.inner == siblingInner {
		return
	}
	if siblingInner.IsAncestor(xmlNode.inner) {
		node.Remove()
		return
	}
	xmlNode.Document.RemoveUnlinkedNode(siblingInner)
	xmlNode.inner.Parent.InsertBefore(siblingInner, xmlNode.inner)
	siblingInner.Valid = true
	return
}

func (xmlNode *XmlNode) addNextSibling(node Node) (err error) {
	nodeType := node.NodeType()
	if nodeType == XML_DOCUMENT_NODE || nodeType == XML_HTML_DOCUMENT_NODE {
		err = ERR_CANNOT_MAKE_DUCMENT_AS_CHILD
		return
	}
	siblingInner := getInternalNode(node)
	if siblingInner == nil || xmlNode.inner == nil || xmlNode.inner.Parent == nil {
		return
	}
	if xmlNode.inner == siblingInner {
		return
	}
	if siblingInner.IsAncestor(xmlNode.inner) {
		node.Remove()
		return
	}
	xmlNode.Document.RemoveUnlinkedNode(siblingInner)
	xmlNode.inner.Parent.InsertAfter(siblingInner, xmlNode.inner)
	siblingInner.Valid = true
	return
}

func (xmlNode *XmlNode) Wrap(data string) (err error) {
	newNodes, err := xmlNode.coerce(data)
	if err == nil && len(newNodes) > 0 {
		newParent := newNodes[0]
		xmlNode.addNextSibling(newParent)
		newParent.AddChild(xmlNode)
	}
	return
}

func (xmlNode *XmlNode) ParseFragment(input, url []byte, options ParseOption) (fragment *DocumentFragment, err error) {
	fragment, err = parsefragment(xmlNode.Document, xmlNode, input, url, options)
	return
}

func (xmlNode *XmlNode) LineNumber() int {
	if xmlNode.inner != nil {
		return xmlNode.inner.Line
	}
	return -1
}

func (xmlNode *XmlNode) RecursivelyRemoveNamespaces() (err error) {
	if xmlNode.inner == nil {
		return
	}
	xmlNode.inner.Ns = nil
	for child := xmlNode.FirstChild(); child != nil; child = child.NextSibling() {
		child.RecursivelyRemoveNamespaces()
	}
	inner := xmlNode.inner
	if (inner.Typ == XML_ELEMENT_NODE || inner.Typ == XML_XINCLUDE_START || inner.Typ == XML_XINCLUDE_END) && len(inner.NsDef) > 0 {
		inner.NsDef = nil
	}
	if inner.Typ == XML_ELEMENT_NODE {
		for _, a := range inner.Props {
			a.Ns = nil
		}
	}
	return
}

func (xmlNode *XmlNode) RemoveDefaultNamespace() {
	if xmlNode.inner == nil {
		return
	}
	inner := xmlNode.inner
	if inner.Ns != nil && inner.Ns.Prefix == "" {
		// Remove from nsDef
		for i, ns := range inner.NsDef {
			if ns == inner.Ns {
				inner.NsDef = append(inner.NsDef[:i], inner.NsDef[i+1:]...)
				break
			}
		}
		inner.Ns = nil
		// Remove from attributes
		for _, a := range inner.Props {
			if a.Ns != nil && a.Ns.Prefix == "" {
				a.Ns = nil
			}
		}
	}
	// Recurse
	for c := inner.Children; c != nil; c = c.Next {
		childNode := newNodeFromInternal(c, xmlNode.Document)
		childNode.RemoveDefaultNamespace()
	}
}

// DeclaredNamespaces returns namespace declarations on this node.
func (xmlNode *XmlNode) DeclaredNamespaces() (result []NamespaceDeclaration) {
	if xmlNode.inner == nil {
		return
	}
	for _, ns := range xmlNode.inner.NsDef {
		result = append(result, NamespaceDeclaration{Prefix: ns.Prefix, Uri: ns.Href})
	}
	return
}

// DeclareNamespace adds a namespace declaration.
func (xmlNode *XmlNode) DeclareNamespace(prefix, href string) {
	if xmlNode.inner == nil || xmlNode.inner.Typ != XML_ELEMENT_NODE {
		return
	}
	// Check if already declared with this href
	existing := xmlNode.inner.LookupNamespaceByHref(href)
	if existing != nil && existing.Prefix == prefix {
		return
	}
	xmlNode.inner.DeclareNamespace(prefix, href)
}

// SetNamespace sets the element's namespace.
func (xmlNode *XmlNode) SetNamespace(prefix, href string) {
	if xmlNode.inner == nil || xmlNode.inner.Typ != XML_ELEMENT_NODE {
		return
	}
	existing := xmlNode.inner.LookupNamespaceByHref(href)
	if existing != nil && existing.Prefix == prefix {
		xmlNode.inner.Ns = existing
		return
	}
	ns := xmlNode.inner.DeclareNamespace(prefix, href)
	xmlNode.inner.Ns = ns
}

// getInternalNode extracts the InternalNode from a Node.
func getInternalNode(n Node) *InternalNode {
	switch v := n.(type) {
	case *XmlNode:
		return v.inner
	case *ElementNode:
		return v.inner
	case *AttributeNode:
		return v.inner
	case *TextNode:
		return v.inner
	case *CDataNode:
		return v.inner
	case *CommentNode:
		return v.inner
	case *ProcessingInstructionNode:
		return v.inner
	}
	return nil
}

// Grow buffer helper
func grow(buffer []byte, n int) (newBuffer []byte) {
	newBuffer = makeSlice(2*cap(buffer) + n)
	copy(newBuffer, buffer)
	return
}

func makeSlice(n int) []byte {
	defer func() {
		if recover() != nil {
			panic(ErrTooLarge)
		}
	}()
	return make([]byte, n)
}

// xmlNameIsValid checks if a string can be used as an XML name.
func xmlNameIsValid(name string) bool {
	if len(name) == 0 {
		return false
	}
	r, sz := utf8.DecodeRuneInString(name)
	if sz == 0 {
		return false
	}
	return isNameStartChar(r)
}

func isNameStartChar(r rune) bool {
	return r == '_' || r == ':' ||
		(r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		r >= 0xC0
}
