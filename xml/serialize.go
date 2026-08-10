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
	enc := xml.NewEncoder(&buf)
	if format&XML_SAVE_FORMAT != 0 {
		enc.Indent("", "  ")
	}

	if format&XML_SAVE_NO_DECL == 0 && root.Doc != nil && root.Doc.DocType == XML_DOCUMENT_NODE {
		encStr := string(encoding)
		if encStr == "" {
			encStr = "utf-8"
		}
		buf.WriteString("<?xml version=\"1.0\" encoding=\"" + encStr + "\"?>\n")
	}

	serializeNodeXML(enc, root, format, 0)
	enc.Flush()
	return buf.Bytes()
}

func serializeNodeXML(enc *xml.Encoder, n *InternalNode, format SerializationOption, depth int) {
	if n == nil {
		return
	}
	switch n.Typ {
	case XML_ELEMENT_NODE:
		// Build start element with namespace
		name := xml.Name{Local: n.Name}
		if n.Ns != nil {
			name.Space = n.Ns.Href
		}

		var attrs []xml.Attr
		// Output namespace declarations
		for _, ns := range n.NsDef {
			if ns.Prefix == "" {
				attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "xmlns"}, Value: ns.Href})
			} else {
				attrs = append(attrs, xml.Attr{Name: xml.Name{Space: "xmlns", Local: ns.Prefix}, Value: ns.Href})
			}
		}
		// Output attributes
		for _, a := range n.Props {
			attrName := xml.Name{Local: a.Name}
			if a.Ns != nil {
				attrName.Space = a.Ns.Href
			}
			attrs = append(attrs, xml.Attr{Name: attrName, Value: a.Value})
		}

		start := xml.StartElement{Name: name, Attr: attrs}
		enc.EncodeToken(start)

		if n.Children != nil {
			for c := n.Children; c != nil; c = c.Next {
				serializeNodeXML(enc, c, format, depth+1)
			}
		}

		enc.EncodeToken(xml.EndElement{Name: name})

	case XML_TEXT_NODE:
		enc.EncodeToken(xml.CharData(n.Content))

	case XML_CDATA_SECTION_NODE:
		// CDATA via directive since encoding/xml doesn't natively support it
		enc.EncodeToken(xml.CharData("<![CDATA[" + n.Content + "]]>"))

	case XML_COMMENT_NODE:
		enc.EncodeToken(xml.Comment([]byte(n.Content)))

	case XML_PI_NODE:
		enc.EncodeToken(xml.ProcInst{Target: n.Name, Inst: []byte(n.Content)})

	case XML_DOCUMENT_NODE, XML_HTML_DOCUMENT_NODE:
		// Document root — serialize children
		for c := n.Children; c != nil; c = c.Next {
			serializeNodeXML(enc, c, format, depth+1)
		}
	}
}

// serializeToHTML serializes an InternalNode subtree to HTML bytes.
func serializeToHTML(root *InternalNode, format SerializationOption, encoding []byte) []byte {
	var buf bytes.Buffer

	if format&XML_SAVE_NO_DECL == 0 && root.Doc != nil {
		// HTML5-style doctype
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

		// Attributes
		for _, a := range n.Props {
			buf.WriteByte(' ')
			buf.WriteString(a.Name)
			buf.WriteString(`="`)
			buf.WriteString(html.EscapeString(a.Value))
			buf.WriteByte('"')
		}

		// Void elements
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
