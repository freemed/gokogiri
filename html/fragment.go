package html

import (
	"bytes"
	"errors"
	"strings"

	"github.com/freemed/gokogiri/xml"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var fragmentWrapperStart = []byte("<div>")
var fragmentWrapperEnd = []byte("</div>")
var fragmentWrapper = []byte("<html><body>")
var bodySigBytes = []byte("<body")

var ErrFailParseFragment = errors.New("failed to parse html fragment")
var ErrEmptyFragment = errors.New("empty html fragment")

const initChildrenNumber = 4

func parsefragment(document xml.Document, node *xml.XmlNode, content, url []byte, options xml.ParseOption) (fragment *xml.DocumentFragment, err error) {
	var root xml.Node

	if node == nil {
		wrapped := append(fragmentWrapper, content...)
		htmlNodes, parseErr := html.ParseFragment(bytes.NewReader(wrapped), &html.Node{
			Type:     html.ElementNode,
			Data:     "body",
			DataAtom: atom.Body,
		})
		if parseErr != nil {
			return nil, ErrFailParseFragment
		}

		// Wrap in a container
		fragDoc := xml.CreateEmptyDocument(document.InputEncoding(), document.OutputEncoding())
		root, err = convertFragmentNodes(fragDoc, htmlNodes)
		if err != nil {
			return nil, err
		}
	} else {
		// Wrapped fragment
		newContent := append(fragmentWrapperStart, content...)
		newContent = append(newContent, fragmentWrapperEnd...)

		htmlNodes, parseErr := html.ParseFragment(bytes.NewReader(newContent), convertToHTMLNode(node))
		if parseErr != nil || len(htmlNodes) == 0 {
			// Try parsing as a full document
			return parsefragment(document, nil, content, url, options)
		}

		fragDoc := xml.CreateEmptyDocument(document.InputEncoding(), document.OutputEncoding())
		root, err = convertFragmentNodes(fragDoc, htmlNodes)
		if err != nil {
			return nil, ErrFailParseFragment
		}
	}

	fragment = &xml.DocumentFragment{Node: root}
	fragment.InEncoding = document.InputEncoding()
	fragment.OutEncoding = document.OutputEncoding()

	document.BookkeepFragment(fragment)
	return
}

func convertToHTMLNode(node *xml.XmlNode) *html.Node {
	nm := node.Name()
	return &html.Node{
		Type:     html.ElementNode,
		Data:     nm,
		DataAtom: atom.Lookup([]byte(nm)),
	}
}

func convertFragmentNodes(doc xml.Document, nodes []*html.Node) (xml.Node, error) {
	if len(nodes) == 0 {
		return nil, ErrEmptyFragment
	}
	var root xml.Node
	for _, n := range nodes {
		gNode := convertHTMLToInternal(n)
		if gNode == nil {
			continue
		}
		xmlNode := xml.NewNode(gNode, doc)
		if root == nil {
			root = xmlNode
		} else {
			xmlNode.AddPreviousSibling(root)
			root = xmlNode
		}
	}
	return root, nil
}

func convertHTMLToInternal(n *html.Node) *xml.InternalNode {
	if n == nil {
		return nil
	}
	var node *xml.InternalNode
	switch n.Type {
	case html.ElementNode:
		node = &xml.InternalNode{
			Typ:   xml.XML_ELEMENT_NODE,
			Name:  n.Data,
			Valid: true,
		}
		for _, attr := range n.Attr {
			a := &xml.InternalAttr{Name: attr.Key, Value: attr.Val}
			if attr.Namespace != "" {
				a.Ns = &xml.InternalNs{Href: attr.Namespace}
			}
			node.Props = append(node.Props, a)
		}
	case html.TextNode:
		node = &xml.InternalNode{
			Typ:     xml.XML_TEXT_NODE,
			Content: n.Data,
			Valid:   true,
		}
	default:
		return nil
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if childNode := convertHTMLToInternal(child); childNode != nil {
			node.AppendChild(childNode)
		}
	}
	return node
}

func ParseFragment(content, inEncoding, url []byte, options xml.ParseOption, outEncoding []byte) (fragment *xml.DocumentFragment, err error) {
	document := CreateEmptyDocument(inEncoding, outEncoding)
	fragment, err = parsefragment(document, nil, content, url, options)
	return
}

func init() {
	_ = strings.TrimSpace
}
