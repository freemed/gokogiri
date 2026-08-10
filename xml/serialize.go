package xml

import (
	"bytes"
	"encoding/xml"
	"html"
	"strings"
)

// serializeToXML serializes an InternalNode subtree to XML bytes.
func serializeToXML(root *InternalNode, format SerializationOption, encoding []byte) []byte {
	var buf bytes.Buffer

	// XML declaration is added at the XmlDocument level (String/ToBuffer)
	serializeNodeXML(&buf, root, format, 0)
	return buf.Bytes()
}

func serializeNodeXML(buf *bytes.Buffer, n *InternalNode, format SerializationOption, depth int) {
	if n == nil {
		return
	}

	indent := ""
	newline := ""
	if format&XML_SAVE_FORMAT != 0 {
		indent = strings.Repeat("  ", depth)
		newline = "\n"
	}

	switch n.Typ {
	case XML_ELEMENT_NODE:
		buf.WriteString(indent)
		buf.WriteByte('<')

		// Write element name with namespace prefix
		if n.Ns != nil && n.Ns.Prefix != "" {
			buf.WriteString(n.Ns.Prefix)
			buf.WriteByte(':')
		}
		buf.WriteString(n.Name)

		// Write namespace declarations
		for _, ns := range n.NsDef {
			buf.WriteByte(' ')
			if ns.Prefix == "" {
				buf.WriteString("xmlns")
			} else {
				buf.WriteString("xmlns:")
				buf.WriteString(ns.Prefix)
			}
			buf.WriteString(`="`)
			xml.EscapeText(buf, []byte(ns.Href))
			buf.WriteByte('"')
		}

		// Write attributes
		for _, a := range n.Props {
			buf.WriteByte(' ')
			if a.Ns != nil && a.Ns.Prefix != "" {
				buf.WriteString(a.Ns.Prefix)
				buf.WriteByte(':')
			}
			buf.WriteString(a.Name)
			buf.WriteString(`="`)
			xml.EscapeText(buf, []byte(a.Value))
			buf.WriteByte('"')
		}

		// Self-closing for empty elements
		if n.Children == nil {
			buf.WriteString("/>")
			if depth > 0 {
				buf.WriteString(newline)
			}
			return
		}

		buf.WriteByte('>')
		buf.WriteString(newline)

		for c := n.Children; c != nil; c = c.Next {
			serializeNodeXML(buf, c, format, depth+1)
		}

		if format&XML_SAVE_FORMAT != 0 && n.Children != nil {
			buf.WriteString(indent)
		}
		buf.WriteString("</")
		if n.Ns != nil && n.Ns.Prefix != "" {
			buf.WriteString(n.Ns.Prefix)
			buf.WriteByte(':')
		}
		buf.WriteString(n.Name)
		buf.WriteByte('>')
		// Only add newline for nested elements (depth > 0, not root)
		if depth > 0 {
			buf.WriteString(newline)
		}

	case XML_TEXT_NODE:
		buf.WriteString(indent)
		// Don't escape tabs for formatted output
		buf.WriteString(n.Content)
		buf.WriteString(newline)

	case XML_CDATA_SECTION_NODE:
		buf.WriteString("<![CDATA[")
		buf.WriteString(n.Content)
		buf.WriteString("]]>")

	case XML_COMMENT_NODE:
		buf.WriteString(indent)
		buf.WriteString("<!--")
		buf.WriteString(n.Content)
		buf.WriteString("-->")
		buf.WriteString(newline)

	case XML_PI_NODE:
		buf.WriteString("<?")
		buf.WriteString(n.Name)
		if n.Content != "" {
			buf.WriteByte(' ')
			buf.WriteString(n.Content)
		}
		buf.WriteString("?>")
		if format&XML_SAVE_FORMAT != 0 {
			buf.WriteString(newline)
		}

	case XML_DOCUMENT_NODE, XML_HTML_DOCUMENT_NODE:
		for c := n.Children; c != nil; c = c.Next {
			serializeNodeXML(buf, c, format, depth)
		}
	}
}

// serializeToHTML serializes an InternalNode subtree to HTML bytes.
func serializeToHTML(root *InternalNode, format SerializationOption, encoding []byte) []byte {
	var buf bytes.Buffer

	if format&XML_SAVE_NO_DECL == 0 && root.Doc != nil {
		buf.WriteString("<!DOCTYPE html>\n")
	}

	serializeNodeHTML(&buf, root, format, 0)
	return buf.Bytes()
}

// HTML void elements that don't have closing tags.
var voidElements = map[string]bool{
	"area": true, "base": true, "br": true, "col": true, "embed": true,
	"hr": true, "img": true, "input": true, "link": true, "meta": true,
	"param": true, "source": true, "track": true, "wbr": true,
}

func serializeNodeHTML(buf *bytes.Buffer, n *InternalNode, format SerializationOption, depth int) {
	if n == nil {
		return
	}
	indent := ""
	if format&XML_SAVE_FORMAT != 0 {
		indent = strings.Repeat("  ", depth)
	}

	switch n.Typ {
	case XML_ELEMENT_NODE:
		buf.WriteString(indent)
		buf.WriteByte('<')
		buf.WriteString(n.Name)

		for _, a := range n.Props {
			buf.WriteByte(' ')
			buf.WriteString(a.Name)
			buf.WriteString(`="`)
			buf.WriteString(html.EscapeString(a.Value))
			buf.WriteByte('"')
		}

		if voidElements[n.Name] {
			buf.WriteString(" />")
			if format&XML_SAVE_FORMAT != 0 {
				buf.WriteByte('\n')
			}
			return
		}

		buf.WriteByte('>')
		if format&XML_SAVE_FORMAT != 0 && n.Children != nil {
			buf.WriteByte('\n')
		}

		for c := n.Children; c != nil; c = c.Next {
			serializeNodeHTML(buf, c, format, depth+1)
		}

		if format&XML_SAVE_FORMAT != 0 && n.Children != nil {
			buf.WriteString(indent)
		}
		buf.WriteString("</")
		buf.WriteString(n.Name)
		buf.WriteByte('>')
		if format&XML_SAVE_FORMAT != 0 {
			buf.WriteByte('\n')
		}

	case XML_TEXT_NODE:
		if format&XML_SAVE_FORMAT != 0 {
			buf.WriteString(indent)
		}
		buf.WriteString(html.EscapeString(n.Content))
		if format&XML_SAVE_FORMAT != 0 {
			buf.WriteByte('\n')
		}

	case XML_CDATA_SECTION_NODE:
		buf.WriteString("<![CDATA[")
		buf.WriteString(n.Content)
		buf.WriteString("]]>")

	case XML_COMMENT_NODE:
		if format&XML_SAVE_FORMAT != 0 {
			buf.WriteString(indent)
		}
		buf.WriteString("<!--")
		buf.WriteString(n.Content)
		buf.WriteString("-->")
		if format&XML_SAVE_FORMAT != 0 {
			buf.WriteByte('\n')
		}

	case XML_PI_NODE:
		buf.WriteString("<?")
		buf.WriteString(n.Name)
		if n.Content != "" {
			buf.WriteByte(' ')
			buf.WriteString(n.Content)
		}
		buf.WriteString("?>")

	case XML_DOCUMENT_NODE, XML_HTML_DOCUMENT_NODE:
		for c := n.Children; c != nil; c = c.Next {
			serializeNodeHTML(buf, c, format, depth)
		}
	}
}
