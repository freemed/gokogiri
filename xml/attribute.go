package xml

/*
AttributeNode represents an attribute, which has a name and a value.

AttributeNodes are created by calling SetAttr or SetNsAttr on an element node,
and retrieved by the Attribute and Attributes functions on an element node.

Note that while namespace declarations resemble attributes, they are a distinct node type
and cannot be used or retrieved as an AttributeNode.
*/
type AttributeNode struct {
	*XmlNode
	sourceAttr *InternalAttr // Write-through back to the element's Props
}

// String returns the value of the attribute.
func (attrNode *AttributeNode) String() string {
	return attrNode.Content()
}

// Value returns the value of the attribute.
func (attrNode *AttributeNode) Value() string {
	return attrNode.Content()
}

// SetValue sets the value of the attribute. Note that the argument will
// be converted to a string, and automatically XML-escaped when the
// document is serialized.
func (attrNode *AttributeNode) SetValue(val interface{}) {
	str := ""
	switch v := val.(type) {
	case string:
		str = v
	default:
		str = ""
	}
	// Write through to the source InternalAttr
	if attrNode.sourceAttr != nil {
		attrNode.sourceAttr.Value = str
	}
	// Also update the wrapper's inner Content
	if attrNode.XmlNode != nil && attrNode.XmlNode.inner != nil {
		attrNode.XmlNode.inner.Content = str
	}
}

// Remove removes this attribute from its parent element.
func (attrNode *AttributeNode) Remove() {
	if attrNode.sourceAttr != nil && attrNode.XmlNode != nil && attrNode.XmlNode.valid {
		// The sourceAttr is a pointer into the element's Props slice.
		// Mark it as removed by setting its value to empty — the actual
		// removal from the Props slice happens when the parent element
		// rebuilds its attribute list. For now, simply zero it.
		attrNode.sourceAttr.Value = ""
		attrNode.sourceAttr.Name = ""
		attrNode.sourceAttr.Ns = nil
	}
	attrNode.XmlNode.valid = false
}
