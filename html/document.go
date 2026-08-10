package html

import (
	"bytes"
	"errors"
	"strings"

	"github.com/freemed/gokogiri/xml"
	"golang.org/x/net/html"
)

// HTML parse options — mirror XML parse option constants.
const (
	HTML_PARSE_RECOVER   xml.ParseOption = 1 << 0
	HTML_PARSE_NODEFDTD  xml.ParseOption = 1 << 2
	HTML_PARSE_NOERROR   xml.ParseOption = 1 << 5
	HTML_PARSE_NOWARNING xml.ParseOption = 1 << 6
	HTML_PARSE_PEDANTIC  xml.ParseOption = 1 << 7
	HTML_PARSE_NOBLANKS  xml.ParseOption = 1 << 8
	HTML_PARSE_NONET     xml.ParseOption = 1 << 11
	HTML_PARSE_NOIMPLIED xml.ParseOption = 1 << 13
	HTML_PARSE_COMPACT   xml.ParseOption = 1 << 16
)

const EmptyHtmlDoc = ""

var DefaultParseOption xml.ParseOption = HTML_PARSE_RECOVER |
	HTML_PARSE_NONET |
	HTML_PARSE_NOERROR |
	HTML_PARSE_NOWARNING

type HtmlDocument struct {
	*xml.XmlDocument
}

var DefaultEncodingBytes = []byte(xml.DefaultEncoding)
var emptyHtmlDocBytes = []byte(EmptyHtmlDoc)

var ErrSetMetaEncoding = errors.New("Set Meta Encoding failed")
var ERR_FAILED_TO_PARSE_HTML = errors.New("failed to parse html input")

// NewDocument creates an HtmlDocument from internal nodes.
func NewDocument(inner *xml.InternalDoc, contentLen int, inEncoding, outEncoding []byte) (doc *HtmlDocument) {
	doc = &HtmlDocument{}
	doc.XmlDocument = xml.NewDocument(inner, contentLen, inEncoding, outEncoding)
	doc.Me = doc
	return
}

// CreateEmptyDocument creates an empty HTML document.
func CreateEmptyDocument(inEncoding, outEncoding []byte) (doc *HtmlDocument) {
	inner := &xml.InternalDoc{
		DocType:       xml.XML_HTML_DOCUMENT_NODE,
		UnlinkedNodes: make(map[*xml.InternalNode]bool),
	}
	doc = NewDocument(inner, 0, inEncoding, outEncoding)
	return
}

// Parse parses HTML content into an HtmlDocument.
func Parse(content, inEncoding, url []byte, options xml.ParseOption, outEncoding []byte) (doc *HtmlDocument, err error) {
	contentLen := len(content)
	if contentLen == 0 {
		return CreateEmptyDocument(inEncoding, outEncoding), nil
	}

	// Use golang.org/x/net/html for parsing
	r := bytes.NewReader(content)
	htmlNode, parseErr := html.Parse(r)
	if parseErr != nil {
		return nil, ERR_FAILED_TO_PARSE_HTML
	}

	// Convert golang.org/x/net/html nodes to our internal nodes
	inEnc := string(inEncoding)
	if inEnc == "" {
		inEnc = "utf-8"
	}
	outEnc := string(outEncoding)
	if outEnc == "" {
		outEnc = "utf-8"
	}

	docRoot := convertHTMLNode(htmlNode)

	// The real root is the first child element (typically <html>)
	// The document wrapper's children contain the HTML structure
	var actualRoot *xml.InternalNode
	if docRoot != nil && docRoot.Children != nil {
		actualRoot = docRoot.Children
		actualRoot.Parent = nil // Detach from document wrapper
	}

	inner := &xml.InternalDoc{
		DocType:       xml.XML_HTML_DOCUMENT_NODE,
		Root:          actualRoot,
		Url:           string(url),
		InEncoding:    inEnc,
		OutEncoding:   outEnc,
		UnlinkedNodes: make(map[*xml.InternalNode]bool),
	}
	if actualRoot != nil {
		actualRoot.Doc = inner
	}

	// Post-processing
	if options&xml.XML_PARSE_NOBLANKS != 0 {
		xml.StripBlankNodes(actualRoot)
	}

	// Build ID index for NodeById lookups
	inner.IDIndex = buildHTMLIDIndex(actualRoot)

	doc = NewDocument(inner, contentLen, []byte(inEnc), []byte(outEnc))
	return
}

// convertHTMLNode recursively converts golang.org/x/net/html nodes to internal nodes.
func convertHTMLNode(n *html.Node) *xml.InternalNode {
	if n == nil {
		return nil
	}

	var node *xml.InternalNode

	switch n.Type {
	case html.ElementNode:
		node = &xml.InternalNode{
			Typ:   xml.XML_ELEMENT_NODE,
			Name:  n.Data,
			Line:  0,
			Valid: true,
		}
		// Attributes
		for _, attr := range n.Attr {
			if attr.Namespace == "" && attr.Key == "xmlns" {
				node.DeclareNamespace("", attr.Val)
			} else if strings.HasPrefix(attr.Key, "xmlns:") {
				node.DeclareNamespace(attr.Key[6:], attr.Val)
			} else {
				a := &xml.InternalAttr{Name: attr.Key, Value: attr.Val}
				if attr.Namespace != "" {
					a.Ns = &xml.InternalNs{Href: attr.Namespace}
				}
				node.Props = append(node.Props, a)
			}
		}

	case html.TextNode:
		node = &xml.InternalNode{
			Typ:     xml.XML_TEXT_NODE,
			Content: n.Data,
			Line:    0,
			Valid:   true,
		}

	case html.CommentNode:
		node = &xml.InternalNode{
			Typ:     xml.XML_COMMENT_NODE,
			Content: n.Data,
			Line:    0,
			Valid:   true,
		}

	case html.DoctypeNode:
		// Doctype is handled as metadata, not stored in tree
		return nil

	case html.DocumentNode:
		node = &xml.InternalNode{
			Typ:   xml.XML_HTML_DOCUMENT_NODE,
			Valid: true,
		}

	default:
		return nil
	}

	// Convert children
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		childNode := convertHTMLNode(child)
		if childNode != nil {
			node.AppendChild(childNode)
		}
	}

	return node
}

// ParseFragment parses an HTML fragment.
func (document *HtmlDocument) ParseFragment(input, url []byte, options xml.ParseOption) (fragment *xml.DocumentFragment, err error) {
	return parsefragment(document, nil, input, url, options)
}

// MetaEncoding extracts the encoding from meta tags.
func (doc *HtmlDocument) MetaEncoding() string {
	root := doc.Root()
	if root == nil {
		return ""
	}
	return findMetaEncoding(root.XmlNode)
}

func findMetaEncoding(xmlNode *xml.XmlNode) string {
	for child := xmlNode.FirstChild(); child != nil; child = child.NextSibling() {
		if child.NodeType() == xml.XML_ELEMENT_NODE {
			if strings.ToLower(child.Name()) == "meta" {
				if strings.ToLower(child.Attr("http-equiv")) == "content-type" {
					content := child.Attr("content")
					for _, part := range strings.Split(content, ";") {
						part = strings.TrimSpace(part)
						if strings.HasPrefix(strings.ToLower(part), "charset=") {
							return part[8:]
						}
					}
				}
				if charset := child.Attr("charset"); charset != "" {
					return charset
				}
			}
		}
		if result := findMetaEncoding(child.(*xml.XmlNode)); result != "" {
			return result
		}
	}
	return ""
}

// SetMetaEncoding sets the encoding in meta tags.
func (doc *HtmlDocument) SetMetaEncoding(encoding string) (err error) {
	root := doc.Root()
	if root == nil {
		return ErrSetMetaEncoding
	}
	err = setMetaEncoding(root.XmlNode, encoding)
	if err != nil {
		return ErrSetMetaEncoding
	}
	return nil
}

func setMetaEncoding(xmlNode *xml.XmlNode, encoding string) error {
	for child := xmlNode.FirstChild(); child != nil; child = child.NextSibling() {
		if child.NodeType() == xml.XML_ELEMENT_NODE {
			if strings.ToLower(child.Name()) == "meta" {
				if strings.ToLower(child.Attr("http-equiv")) == "content-type" {
					child.SetAttr("content", "text/html; charset="+encoding)
					return nil
				}
				if child.Attr("charset") != "" {
					child.SetAttr("charset", encoding)
					return nil
				}
			}
		}
		if err := setMetaEncoding(child.(*xml.XmlNode), encoding); err == nil {
			return nil
		}
	}
	return ErrSetMetaEncoding
}

// buildHTMLIDIndex walks the tree and indexes elements with id attributes.
func buildHTMLIDIndex(root *xml.InternalNode) map[string]*xml.InternalNode {
	if root == nil {
		return nil
	}
	index := make(map[string]*xml.InternalNode)
	var walk func(n *xml.InternalNode)
	walk = func(n *xml.InternalNode) {
		if n.Typ == xml.XML_ELEMENT_NODE {
			for _, a := range n.Props {
				if strings.ToLower(a.Name) == "id" && a.Ns == nil && a.Value != "" {
					if _, exists := index[a.Value]; !exists {
						index[a.Value] = n
					}
				}
			}
		}
		for c := n.Children; c != nil; c = c.Next {
			walk(c)
		}
	}
	walk(root)
	if len(index) == 0 {
		return nil
	}
	return index
}
