package xpath

import (
	antchfx "github.com/antchfx/xpath"
)

// NodeAdapter provides the minimal DOM navigation interface needed for XPath.
// Implemented by xml.InternalNode to bridge between our DOM and antchfx/xpath.
type NodeAdapter interface {
	XPathType() int
	XPathName() string
	XPathValue() string
	XPathPrefix() string
	XPathNamespaceURI() string
	XPathParent() interface{}    // returns NodeAdapter or nil
	XPathFirstChild() interface{} // returns NodeAdapter or nil
	XPathNextSibling() interface{} // returns NodeAdapter or nil
	XPathPrevSibling() interface{} // returns NodeAdapter or nil
	XPathAttributes() interface{}  // returns []AttrAdapter or compatible slice
	XPathCopy() interface{}        // returns NodeAdapter
}

// AttrAdapter holds attribute info for XPath navigation.
type AttrAdapter struct {
	Name         string
	Value        string
	Prefix       string
	NamespaceURI string
}

// nodeNavigator implements antchfx.NodeNavigator for gokogiri DOM nodes.
type nodeNavigator struct {
	node    NodeAdapter
	attrIdx int
	attrs   []AttrAdapter
}

// NewNavigator creates a new antchfx.NodeNavigator from a gokogiri DOM node.
func NewNavigator(node NodeAdapter) antchfx.NodeNavigator {
	return &nodeNavigator{node: node, attrIdx: -1}
}

func (n *nodeNavigator) NodeType() antchfx.NodeType {
	if n.node == nil {
		return antchfx.RootNode
	}
	if n.attrIdx >= 0 && n.attrIdx < len(n.attrs) {
		return antchfx.AttributeNode
	}
	return antchfx.NodeType(n.node.XPathType())
}

func (n *nodeNavigator) LocalName() string {
	if n.node == nil {
		return ""
	}
	if n.attrIdx >= 0 && n.attrIdx < len(n.attrs) {
		return n.attrs[n.attrIdx].Name
	}
	return n.node.XPathName()
}

func (n *nodeNavigator) Prefix() string {
	if n.node == nil {
		return ""
	}
	if n.attrIdx >= 0 && n.attrIdx < len(n.attrs) {
		return n.attrs[n.attrIdx].Prefix
	}
	return n.node.XPathPrefix()
}

func (n *nodeNavigator) Value() string {
	if n.node == nil {
		return ""
	}
	if n.attrIdx >= 0 && n.attrIdx < len(n.attrs) {
		return n.attrs[n.attrIdx].Value
	}
	return n.node.XPathValue()
}

func (n *nodeNavigator) NamespaceURI() string {
	if n.node == nil {
		return ""
	}
	if n.attrIdx >= 0 && n.attrIdx < len(n.attrs) {
		return n.attrs[n.attrIdx].NamespaceURI
	}
	return n.node.XPathNamespaceURI()
}

func (n *nodeNavigator) Copy() antchfx.NodeNavigator {
	if n.node == nil {
		return &nodeNavigator{node: nil, attrIdx: -1}
	}
	// Return a navigator that wraps the SAME node. antchfx/xpath 
	// uses Copy for snapshotting positions, and a deep copy of the
	// entire DOM subtree is expensive and can produce invalid pointers.
	return &nodeNavigator{node: n.node, attrIdx: -1}
}

func (n *nodeNavigator) MoveToRoot() {
	if n.node == nil {
		return
	}
	for {
		p := n.node.XPathParent()
		if p == nil {
			break
		}
		if a, ok := p.(NodeAdapter); ok {
			n.node = a
		} else {
			break
		}
	}
	n.attrIdx = -1
	n.attrs = nil
}

func (n *nodeNavigator) MoveToParent() bool {
	if n.attrIdx >= 0 {
		// When on an attribute, parent is the owning element
		n.attrIdx = -1
		n.attrs = nil
		return true
	}
	p := n.node.XPathParent()
	if p == nil {
		return false
	}
	if a, ok := p.(NodeAdapter); ok {
		n.node = a
		n.attrIdx = -1
		n.attrs = nil
		return true
	}
	return false
}

func (n *nodeNavigator) MoveToNextAttribute() bool {
	if n.attrs == nil {
		raw := n.node.XPathAttributes()
		n.attrs = attrsFromRaw(raw)
		n.attrIdx = -1
	}
	n.attrIdx++
	if n.attrIdx >= len(n.attrs) {
		return false
	}
	return true
}

// attrsFromRaw decodes attribute data from various slice types.
func attrsFromRaw(raw interface{}) []AttrAdapter {
	switch v := raw.(type) {
	case []AttrAdapter:
		return v
	case []interface{}:
		var out []AttrAdapter
		for _, item := range v {
			if a, ok := item.(AttrAdapter); ok {
				out = append(out, a)
			} else if m, ok := item.(map[string]interface{}); ok {
				a := AttrAdapter{}
				if name, ok := m["Name"].(string); ok {
					a.Name = name
				}
				if val, ok := m["Value"].(string); ok {
					a.Value = val
				}
				if prefix, ok := m["Prefix"].(string); ok {
					a.Prefix = prefix
				}
				if ns, ok := m["NamespaceURI"].(string); ok {
					a.NamespaceURI = ns
				}
				out = append(out, a)
			}
		}
		return out
	}
	// The xml package returns []xml.xpathAttr — we handle this via
	// a known interface method. If we got here, attrs are empty.
	return nil
}

func (n *nodeNavigator) MoveToChild() bool {
	c := n.node.XPathFirstChild()
	if c == nil {
		return false
	}
	if a, ok := c.(NodeAdapter); ok {
		n.node = a
		n.attrIdx = -1
		n.attrs = nil
		return true
	}
	return false
}

func (n *nodeNavigator) MoveToFirst() bool {
	for {
		p := n.node.XPathPrevSibling()
		if p == nil {
			return true
		}
		if a, ok := p.(NodeAdapter); ok {
			n.node = a
		} else {
			return true
		}
	}
}

func (n *nodeNavigator) MoveToNext() bool {
	s := n.node.XPathNextSibling()
	if s == nil {
		return false
	}
	if a, ok := s.(NodeAdapter); ok {
		n.node = a
		n.attrIdx = -1
		n.attrs = nil
		return true
	}
	return false
}

func (n *nodeNavigator) MoveToPrevious() bool {
	s := n.node.XPathPrevSibling()
	if s == nil {
		return false
	}
	if a, ok := s.(NodeAdapter); ok {
		n.node = a
		n.attrIdx = -1
		n.attrs = nil
		return true
	}
	return false
}

func (n *nodeNavigator) MoveTo(other antchfx.NodeNavigator) bool {
	otherNav, ok := other.(*nodeNavigator)
	if !ok {
		return false
	}
	n.node = otherNav.node
	n.attrIdx = otherNav.attrIdx
	n.attrs = otherNav.attrs
	return true
}

// resultNode returns the current node as a NodeAdapter, creating an AttrNode
// if the navigator is positioned on an attribute.
func (n *nodeNavigator) resultNode() NodeAdapter {
	if n.attrIdx >= 0 && n.attrIdx < len(n.attrs) {
		a := n.attrs[n.attrIdx]
		return &AttrNode{
			Name_:         a.Name,
			Value_:        a.Value,
			Prefix_:       a.Prefix,
			NamespaceURI_: a.NamespaceURI,
		}
	}
	return n.node
}
