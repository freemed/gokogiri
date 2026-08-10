package xml

// TextNode represents a text node in the XML tree.
type TextNode struct {
	*XmlNode
}

// DisableOutputEscaping is a no-op in the pure Go implementation.
// It previously disabled libxml2's entity escaping for '<', '>', and '&'
// characters. The pure Go serializer already handles content faithfully.
func (node *TextNode) DisableOutputEscaping() {
	// No-op: Go serializers don't have a separate escaping toggle
}
