package xml

// InternalNode is the concrete tree node in the pure-Go DOM.
type InternalNode struct {
	Typ      NodeType
	Name     string // local name (no prefix)
	Content  string // text content (text/cdata/comment/pi nodes)
	Parent   *InternalNode
	Children *InternalNode // first child
	Last     *InternalNode // last child
	Prev     *InternalNode
	Next     *InternalNode
	Props    []*InternalAttr
	Ns       *InternalNs
	NsDef    []*InternalNs
	Doc      *InternalDoc
	Line     int
	Valid    bool
}

// previousSiblings returns previous siblings ordered nearest to furthest.
func (n *InternalNode) previousSiblings() []*InternalNode {
	var out []*InternalNode
	for p := n.Prev; p != nil; p = p.Prev {
		out = append(out, p)
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

// InternalAttr represents an XML attribute.
type InternalAttr struct {
	Name  string
	Value string
	Ns    *InternalNs
}

// InternalNs represents a namespace.
type InternalNs struct {
	Prefix string
	Href   string
}

// DTDInfo holds metadata extracted during DTD pre-scan.
type DTDInfo struct {
	Entities     map[string]DTDEntity
	IDAttrs      map[string]string
	DefaultAttrs map[string]map[string]string
	Notations    map[string]string
}

// DTDEntity represents a parsed DTD entity declaration.
type DTDEntity struct {
	Value    string
	SystemID string
	PublicID string
	NDATA    string
	IsNDATA  bool
}

// InternalDoc is the owning document for a DOM tree.
type InternalDoc struct {
	Root          *InternalNode
	Url           string
	InEncoding    string
	OutEncoding   string
	DocType       NodeType
	UnlinkedNodes map[*InternalNode]bool
	Fragments     []*DocumentFragment
	IDIndex       map[string]*InternalNode
	DTDInfo       *DTDInfo
}

// ---- Tree manipulation methods ----

// Detach removes this node from its parent's child list.
func (n *InternalNode) Detach() {
	if n.Parent == nil {
		return
	}
	if n.Parent.Children == n {
		n.Parent.Children = n.Next
	}
	if n.Parent.Last == n {
		n.Parent.Last = n.Prev
	}
	if n.Prev != nil {
		n.Prev.Next = n.Next
	}
	if n.Next != nil {
		n.Next.Prev = n.Prev
	}
	n.Parent = nil
	n.Prev = nil
	n.Next = nil
}

// AppendChild adds child as the last child of n.
func (n *InternalNode) AppendChild(child *InternalNode) {
	child.Detach()
	child.Parent = n
	if n.Last != nil {
		n.Last.Next = child
		child.Prev = n.Last
	} else {
		n.Children = child
	}
	n.Last = child
}

// PrependChild adds child as the first child of n.
func (n *InternalNode) PrependChild(child *InternalNode) {
	child.Detach()
	child.Parent = n
	child.Next = n.Children
	if n.Children != nil {
		n.Children.Prev = child
	} else {
		n.Last = child
	}
	n.Children = child
}

// InsertBefore inserts child before ref in n's child list.
func (n *InternalNode) InsertBefore(child, ref *InternalNode) {
	if ref == nil || ref.Parent != n {
		n.AppendChild(child)
		return
	}
	child.Detach()
	child.Parent = n
	child.Prev = ref.Prev
	child.Next = ref
	if ref.Prev != nil {
		ref.Prev.Next = child
	} else {
		n.Children = child
	}
	ref.Prev = child
}

// InsertAfter inserts child after ref in n's child list.
func (n *InternalNode) InsertAfter(child, ref *InternalNode) {
	if ref == nil || ref.Parent != n {
		n.AppendChild(child)
		return
	}
	child.Detach()
	child.Parent = n
	child.Prev = ref
	child.Next = ref.Next
	if ref.Next != nil {
		ref.Next.Prev = child
	} else {
		n.Last = child
	}
	ref.Next = child
}

// RemoveChildren detaches all children.
func (n *InternalNode) RemoveChildren() []*InternalNode {
	var out []*InternalNode
	for c := n.Children; c != nil; c = c.Next {
		out = append(out, c)
	}
	for _, c := range out {
		c.Parent = nil
		c.Prev = nil
		c.Next = nil
	}
	n.Children = nil
	n.Last = nil
	return out
}

// IsAncestor returns true if other is an ancestor of n.
func (n *InternalNode) IsAncestor(other *InternalNode) bool {
	for p := n.Parent; p != nil; p = p.Parent {
		if p == other {
			return true
		}
	}
	return false
}

// DeepCopy creates a recursive copy.
func (n *InternalNode) DeepCopy(targetDoc *InternalDoc, level int) *InternalNode {
	cp := &InternalNode{
		Typ:   n.Typ,
		Name:  n.Name,
		Content: n.Content,
		Doc:   targetDoc,
		Line:  n.Line,
		Valid: true,
	}
	if n.Ns != nil {
		cp.Ns = &InternalNs{Prefix: n.Ns.Prefix, Href: n.Ns.Href}
	}
	for _, nsd := range n.NsDef {
		cp.NsDef = append(cp.NsDef, &InternalNs{Prefix: nsd.Prefix, Href: nsd.Href})
	}
	for _, a := range n.Props {
		ca := &InternalAttr{Name: a.Name, Value: a.Value}
		if a.Ns != nil {
			ca.Ns = &InternalNs{Prefix: a.Ns.Prefix, Href: a.Ns.Href}
		}
		cp.Props = append(cp.Props, ca)
	}
	if level > 0 {
		for c := n.Children; c != nil; c = c.Next {
			childCopy := c.DeepCopy(targetDoc, level-1)
			cp.AppendChild(childCopy)
		}
	}
	return cp
}

// SetAttr sets a non-namespaced attribute.
func (n *InternalNode) SetAttr(name, value string) {
	for _, a := range n.Props {
		if a.Ns == nil && a.Name == name {
			a.Value = value
			return
		}
	}
	n.Props = append(n.Props, &InternalAttr{Name: name, Value: value})
}

// SetNsAttr sets a namespaced attribute.
func (n *InternalNode) SetNsAttr(href, name, value string) {
	for _, a := range n.Props {
		if a.Ns != nil && a.Ns.Href == href && a.Name == name {
			a.Value = value
			return
		}
	}
	n.Props = append(n.Props, &InternalAttr{
		Name:  name,
		Value: value,
		Ns:    &InternalNs{Href: href},
	})
}

// GetAttr returns an attribute value by name.
func (n *InternalNode) GetAttr(name string) (string, bool) {
	for _, a := range n.Props {
		if a.Ns == nil && a.Name == name {
			return a.Value, true
		}
	}
	return "", false
}

// DeclareNamespace adds a namespace declaration.
func (n *InternalNode) DeclareNamespace(prefix, href string) *InternalNs {
	for _, ns := range n.NsDef {
		if ns.Prefix == prefix {
			if ns.Href == href {
				return ns
			}
			ns.Href = href
			return ns
		}
	}
	ns := &InternalNs{Prefix: prefix, Href: href}
	n.NsDef = append(n.NsDef, ns)
	return ns
}

// LookupNamespace finds a namespace by prefix walking up the tree.
func (n *InternalNode) LookupNamespace(prefix string) *InternalNs {
	for cur := n; cur != nil; cur = cur.Parent {
		for _, ns := range cur.NsDef {
			if ns.Prefix == prefix {
				return ns
			}
		}
	}
	return nil
}

// LookupNamespaceByHref finds a namespace by href walking up the tree.
func (n *InternalNode) LookupNamespaceByHref(href string) *InternalNs {
	for cur := n; cur != nil; cur = cur.Parent {
		for _, ns := range cur.NsDef {
			if ns.Href == href {
				return ns
			}
		}
	}
	return nil
}

// CountChildElements returns the number of element children.
func (n *InternalNode) CountChildElements() int {
	count := 0
	for c := n.Children; c != nil; c = c.Next {
		if c.Typ == XML_ELEMENT_NODE {
			count++
		}
	}
	return count
}
