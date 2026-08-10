package xml

// Nodeset is a slice of nodes, used for XPath result handling.
type Nodeset []Node

// ToPointers returns the internal pointers for each node.
func (n Nodeset) ToPointers() (pointers []interface{}) {
	for _, node := range n {
		pointers = append(pointers, node.NodePtr())
	}
	return
}

// ToXPathNodeset returns a representation suitable for XPath interop.
// Phase 4 will implement full antchfx/xpath integration.
func (n Nodeset) ToXPathNodeset() interface{} {
	return n
}

// ToXPathValueTree returns a representation as a value tree.
// Phase 4 will implement full antchfx/xpath integration.
func (n Nodeset) ToXPathValueTree() interface{} {
	return n
}
