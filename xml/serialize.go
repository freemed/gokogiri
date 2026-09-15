package xml

import (
	"bytes"
	"encoding/xml"
	"html"
	"strings"
)

// serializeToXML serializes an InternalNode subtree to XML bytes.
//
// This is the DEFAULT path and it preserves gokogiri's historical output byte
// for byte: with XML_SAVE_FORMAT each child is written on its own line, text
// nodes are written verbatim (unescaped) and attribute values are escaped with
// encoding/xml.EscapeText. The reference fixtures under xml/tests/ are the
// expectation for this path, so nothing here may change without changing those
// fixtures - which is exactly what must not happen.
//
// Callers that need libxml2/libxslt-compatible output (the XSLT engine, whose
// result is compared against xsltproc) opt in with XML_SAVE_LIBXSLT, which
// selects the libxml2-faithful serializer further down this file.
func serializeToXML(root *InternalNode, format SerializationOption, encoding []byte) []byte {
	if format&XML_SAVE_LIBXSLT != 0 {
		return serializeToXMLLibxslt(root, format, encoding)
	}

	var buf bytes.Buffer

	// XML declaration is added at the XmlDocument level (String/ToBuffer)
	serializeNodeXML(&buf, root, format, 0)
	return buf.Bytes()
}

// serializeToXMLLibxslt serializes a subtree the way libxml2/libxslt does, so
// that XSLT output matches xsltproc:
//
//   - Whitespace (newline + one two-space indent step per level) is inserted
//     only around the element/comment/PI children of an element, and before
//     that element's end tag. An element with any text or CDATA child is
//     "mixed content" and is written completely verbatim, so indentation is
//     never injected into a text node.
//   - Text content is escaped as libxml2 does it: & < > become entities and a
//     carriage return becomes &#13;. A raw CR would be normalized to a
//     linefeed by any conformant parser, so it has to be escaped to survive a
//     serialize/parse round trip (this matters for X12 CRLF terminators).
//   - Attribute values escape & < > " plus tab/LF/CR as character references.
//   - XML_SAVE_AS_TEXT (the XSLT "text" output method) writes characters
//     verbatim, without either escaping or formatting.
//
// Only reached when XML_SAVE_LIBXSLT is set.
func serializeToXMLLibxslt(root *InternalNode, format SerializationOption, encoding []byte) []byte {
	var buf bytes.Buffer

	indent := format&XML_SAVE_FORMAT != 0 && format&XML_SAVE_AS_TEXT == 0
	escape := format&XML_SAVE_AS_TEXT == 0
	libxsltSerializeNode(&buf, root, indent, escape, 0)
	return buf.Bytes()
}

// serializeNodeXML writes one node using the default (historical) format.
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

// hasTextContent reports whether n has any text-like child. Elements with
// text/CDATA children are mixed content: libxml2 does not add indentation
// whitespace inside them.
func hasTextContent(n *InternalNode) bool {
	for c := n.Children; c != nil; c = c.Next {
		switch c.Typ {
		case XML_TEXT_NODE, XML_CDATA_SECTION_NODE:
			return true
		}
	}
	return false
}

// writeIndent writes depth levels of the two-space indentation step used by
// libxml2's xmlTreeIndentString.
func writeIndent(buf *bytes.Buffer, depth int) {
	for i := 0; i < depth; i++ {
		buf.WriteString("  ")
	}
}

// escapeXMLText writes s using libxml2's text-node escaping: & < > as
// entities, CR as a decimal character reference, everything else (including
// tabs, linefeeds, quotes and apostrophes) verbatim.
func escapeXMLText(buf *bytes.Buffer, s string) {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '&':
			buf.WriteString("&amp;")
		case '<':
			buf.WriteString("&lt;")
		case '>':
			buf.WriteString("&gt;")
		case '\r':
			buf.WriteString("&#13;")
		default:
			buf.WriteByte(s[i])
		}
	}
}

// escapeXMLAttr writes s using libxml2's attribute-value escaping: & < " as
// entities plus tab/LF/CR as decimal character references so that the XML
// attribute-value normalization rules cannot lose them.
func escapeXMLAttr(buf *bytes.Buffer, s string) {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '&':
			buf.WriteString("&amp;")
		case '<':
			buf.WriteString("&lt;")
		case '"':
			buf.WriteString("&quot;")
		case '>':
			buf.WriteString("&gt;")
		case '\t':
			buf.WriteString("&#9;")
		case '\n':
			buf.WriteString("&#10;")
		case '\r':
			buf.WriteString("&#13;")
		default:
			buf.WriteByte(s[i])
		}
	}
}

// writeXMLTag opens or closes an element name, including its namespace prefix.
func writeXMLName(buf *bytes.Buffer, n *InternalNode) {
	if n.Ns != nil && n.Ns.Prefix != "" {
		buf.WriteString(n.Ns.Prefix)
		buf.WriteByte(':')
	}
	buf.WriteString(n.Name)
}

// serializeNodeXML writes a single node. Callers are responsible for any
// leading indentation: this writes the node itself and, for block-formatted
// elements, the newlines and indentation of its children.
func libxsltSerializeNode(buf *bytes.Buffer, n *InternalNode, indent, escape bool, depth int) {
	if n == nil {
		return
	}

	switch n.Typ {
	case XML_ELEMENT_NODE:
		buf.WriteByte('<')
		writeXMLName(buf, n)

		// Namespace declarations
		for _, ns := range n.NsDef {
			buf.WriteByte(' ')
			if ns.Prefix == "" {
				buf.WriteString("xmlns")
			} else {
				buf.WriteString("xmlns:")
				buf.WriteString(ns.Prefix)
			}
			buf.WriteString(`="`)
			escapeXMLAttr(buf, ns.Href)
			buf.WriteByte('"')
		}

		// Attributes
		for _, a := range n.Props {
			buf.WriteByte(' ')
			if a.Ns != nil && a.Ns.Prefix != "" {
				buf.WriteString(a.Ns.Prefix)
				buf.WriteByte(':')
			}
			buf.WriteString(a.Name)
			buf.WriteString(`="`)
			escapeXMLAttr(buf, a.Value)
			buf.WriteByte('"')
		}

		// Self-closing for empty elements
		if n.Children == nil {
			buf.WriteString("/>")
			return
		}

		buf.WriteByte('>')

		// Block formatting only when there is no text content to protect.
		block := indent && !hasTextContent(n)
		for c := n.Children; c != nil; c = c.Next {
			if block {
				buf.WriteByte('\n')
				writeIndent(buf, depth+1)
			}
			libxsltSerializeNode(buf, c, indent, escape, depth+1)
		}
		if block {
			buf.WriteByte('\n')
			writeIndent(buf, depth)
		}

		buf.WriteString("</")
		writeXMLName(buf, n)
		buf.WriteByte('>')

	case XML_TEXT_NODE:
		if escape {
			escapeXMLText(buf, n.Content)
		} else {
			buf.WriteString(n.Content)
		}

	case XML_CDATA_SECTION_NODE:
		buf.WriteString("<![CDATA[")
		buf.WriteString(n.Content)
		buf.WriteString("]]>")

	case XML_COMMENT_NODE:
		buf.WriteString("<!--")
		buf.WriteString(n.Content)
		buf.WriteString("-->")

	case XML_PI_NODE:
		buf.WriteString("<?")
		buf.WriteString(n.Name)
		if n.Content != "" {
			buf.WriteByte(' ')
			buf.WriteString(n.Content)
		}
		buf.WriteString("?>")

	case XML_DOCUMENT_NODE, XML_HTML_DOCUMENT_NODE:
		// Document-level children are written back to back: libxml2 adds no
		// whitespace between them.
		for c := n.Children; c != nil; c = c.Next {
			libxsltSerializeNode(buf, c, indent, escape, depth)
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
